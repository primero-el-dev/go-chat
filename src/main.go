package main

//func main() {
//	_, err := http.Get("https://jsonplaceholder.typicode.com/posts/1")
//	if err != nil {
//		log.Fatalln(err)
//	}
//}

//func main() {
//	app := app.New()
//	window := app.NewWindow("Hello World")
//
//	loginText := canvas.NewText("Login", color.White)
//	loginInput := widget.NewEntry()
//	row := container.NewHBox(loginText, loginInput)
//	content := container.NewVBox(row)
//	window.SetContent(content)
//
//	//hello := widget.NewLabel("0")
//	//window.SetContent(container.NewVBox(
//	//	hello,
//	//	widget.NewForm(
//	//		widget.Gri),
//	//	widget.NewButton("2", func() {
//	//		value, err := strconv.Atoi(hello.Text)
//	//		if err != nil {
//	//			panic(err)
//	//		}
//	//		hello.SetText(strconv.Itoa(value + 2))
//	//	}),
//	//))
//	window.ShowAndRun()
//}

//
//type Calculation struct {
//	CurrentNumber  string
//	PreviousNumber string
//	PreviousAction string
//}
//
//func (calc *Calculation) WriteChar(char string) {
//	_, err := strconv.Atoi(char)
//	if err == nil {
//		calc.CurrentNumber += char
//	} else if char == "." && !strings.Contains(calc.CurrentNumber, char) {
//		calc.CurrentNumber += char
//	}
//}
