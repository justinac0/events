package main

import (
	"events/internal"
	"events/web/templates"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/a-h/templ"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

const (
	UndefinedEvents string = "UNDEFINED_EVENTS"
	UIEvents        string = "UI_EVENTS"
)

type AppState struct {
	Aborted  bool
	Testing  bool
	Numbers  chan int
	EventBus *internal.EventBus
}

func NewAppState() *AppState {
	return &AppState{
		Aborted:  false,
		Testing:  false,
		Numbers:  make(chan int),
		EventBus: internal.NewEventBus(),
	}
}

var State = NewAppState()

func Render(ctx echo.Context, statusCode int, t templ.Component) error {
	ctx.Response().Writer.WriteHeader(statusCode)
	ctx.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTML)
	return t.Render(ctx.Request().Context(), ctx.Response().Writer)
}

func handleUIEvents(eb *internal.EventBus) {
	ch := make(chan interface{}, 16)
	eb.Subscribe(UIEvents, ch)

	for {
		data := <-ch
		fmt.Printf("[%v]: %v\n", UIEvents, data)
		switch data {
		case "abort":
			State.Aborted = true
		case "start":
			go startTests()
		}
	}
}

func handleUndefinedEvents(eb *internal.EventBus) {
	ch := make(chan interface{}, 16)
	eb.Subscribe(UndefinedEvents, ch)

	for {
		data := <-ch
		fmt.Printf("[%v]: %v\n", UndefinedEvents, data)
	}
}

func startTests() {
	State.Aborted = false
	State.Testing = true

	i := 0
	for i < 100 && !State.Aborted {
		time.Sleep(1 * time.Second)
		State.Numbers <- i + 1
		i++
	}

	State.Testing = false
}

func main() {
	e := echo.New()
	e.Static("css", "web/static/css")

	e.HideBanner = true

	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "method=${method}, uri=${uri}, status=${status}\n",
	}))

	e.GET("/", func(c echo.Context) error {
		return Render(c, http.StatusOK, templates.Index())
	})

	e.GET("/poll-results", func(c echo.Context) error {
		number := <-State.Numbers
		return c.String(http.StatusOK, "<h2 id='results' hx-swap-oob='true'>"+strconv.Itoa(number)+"</h2>")
	})

	e.POST("/start", func(c echo.Context) error {
		if State.Testing {
			return c.NoContent(http.StatusOK)
		}

		State.EventBus.Publish(UIEvents, "start")

		return c.NoContent(http.StatusOK)
	})

	e.POST("/send-event", func(c echo.Context) error {
		eventName := c.FormValue("event")
		switch eventName {
		case "abort":
			State.EventBus.Publish(UIEvents, eventName)
		default:
			State.EventBus.Publish(UndefinedEvents, eventName)
		}

		return c.NoContent(http.StatusOK)
	})

	go handleUIEvents(State.EventBus)
	go handleUndefinedEvents(State.EventBus)

	e.Logger.Fatal(e.Start("127.0.0.1:80"))
}
