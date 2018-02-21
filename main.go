package main

import (
	"github.com/dtylman/gowd"

	"fmt"
	"time"

	"github.com/dtylman/gowd/bootstrap"
)

var body *gowd.Element

func main() {
	//creates a new bootstrap fluid container
	body = bootstrap.NewContainer(false)
	// add some elements using the object model
	div := bootstrap.NewElement("div", "well")
	row := bootstrap.NewRow(bootstrap.NewColumn(bootstrap.ColumnSmall, 3, div))
	body.AddElement(row)
	// add some other elements from HTML
	div.AddHTML(`<div class="dropdown">
	<button class="btn btn-primary dropdown-toggle" type="button" data-toggle="dropdown">Prueba increíble
	<span class="caret"></span></button>
	<ul class="dropdown-menu" id="dropdown-menu">
	<li><a href="#">HTML</a></li>
	<li><a href="#">CSS</a></li>
	<li><a href="#">JavaScript</a></li>
	<li><a href="#">Esto no sirve para nada</a></li>
	</ul>
	</div>`, nil)
	// add a button to show a progress bar
	btn := bootstrap.NewButton(bootstrap.ButtonPrimary, "Vamoooo")
	btn.OnEvent(gowd.OnClick, btnClicked)
	row.AddElement(bootstrap.NewColumn(bootstrap.ColumnSmall, 3, bootstrap.NewElement("div", "well", btn)))

	//start the ui loop
	gowd.Run(body)
}

// happens when the 'start' button is clicked
func btnClicked(sender *gowd.Element, event *gowd.EventElement) {
	// adds a text and progress bar to the body
	sender.SetText("Calma...")
	text := body.AddElement(gowd.NewStyledText("Lets'go...", gowd.BoldText))
	progressBar := bootstrap.NewProgressBar()
	body.AddElement(progressBar.Element)

	// makes the body stop responding to user events
	body.Disable()

	// clean up - remove the added elements
	defer func() {
		sender.SetText("Vamoooo")
		body.RemoveElement(text)
		body.RemoveElement(progressBar.Element)
		body.Enable()
	}()

	// render the progress bar
	for i := 0; i <= 100; i++ {
		progressBar.SetValue(i, 123)
		text.SetText(fmt.Sprintf("Enviando dinero a tu cuenta %v", i))
		time.Sleep(time.Millisecond * 20)
		// this will cause the body to be refreshed
		body.Render()
	}

}
