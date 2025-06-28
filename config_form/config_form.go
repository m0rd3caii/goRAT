package main

import (
	"fmt"
	"image/color"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

const configFilePath = "config/config.go"
const backgroundImagePath = "config_form/back.png"
const logoImagePath = "./bg.jpg"
const titleImagePath = "./bg.jpg"
const tokenLabelImagePath = "./bg.jpg"
const channelLabelImagePath = "./bg.jpg"

// CustomTheme extends the default theme
type CustomTheme struct {
	fyne.Theme
}

func newCustomTheme() fyne.Theme {
	return &CustomTheme{Theme: theme.DefaultTheme()}
}

func (t *CustomTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNamePrimary:
		return color.NRGBA{R: 65, G: 105, B: 225, A: 255} // Royal blue
	case theme.ColorNameBackground:
		return color.NRGBA{R: 240, G: 248, B: 255, A: 255} // Very light blue
	case theme.ColorNameButton:
		return color.NRGBA{R: 70, G: 130, B: 180, A: 255} // Steel blue
	default:
		return t.Theme.Color(name, variant)
	}
}

func form() {
	myApp := app.New()
	myApp.Settings().SetTheme(newCustomTheme())
	myWindow := myApp.NewWindow("goRat Configuration")
	myWindow.Resize(fyne.NewSize(900, 600))
	myWindow.CenterOnScreen()

	// Logo and title
	var logo *canvas.Image
	if _, err := os.Stat(logoImagePath); err == nil {
		absLogoPath, _ := filepath.Abs(logoImagePath)
		logo = canvas.NewImageFromFile(absLogoPath)
		logo.SetMinSize(fyne.NewSize(100, 100))
		logo.FillMode = canvas.ImageFillContain
	}

	// Title image or fallback to text title
	var titleContent fyne.CanvasObject
	if _, err := os.Stat(titleImagePath); err == nil {
		absTitlePath, _ := filepath.Abs(titleImagePath)
		titleImage := canvas.NewImageFromFile(absTitlePath)
		titleImage.SetMinSize(fyne.NewSize(300, 60)) // Adjust size as needed
		titleImage.FillMode = canvas.ImageFillContain
		titleContent = titleImage
	} else {
		// Fallback to text title if image not found
		textTitle := canvas.NewText("goRAT", color.NRGBA{R: 25, G: 25, B: 112, A: 255})
		textTitle.Alignment = fyne.TextAlignCenter
		textTitle.TextSize = 32
		textTitle.TextStyle = fyne.TextStyle{Bold: true}
		titleContent = textTitle
		fmt.Printf("Warning: No title image found at path: %s, falling back to text title\n", titleImagePath)
	}

	subtitle := canvas.NewText("Configure your bot settings below", color.NRGBA{R: 70, G: 70, B: 70, A: 255})
	subtitle.Alignment = fyne.TextAlignCenter
	subtitle.TextSize = 16

	// Create styled entries directly with the entries accessible
	botTokenEntry := widget.NewEntry()
	botTokenEntry.SetPlaceHolder("Enter your bot token here...")

	ServerIDEntry := widget.NewEntry()
	ServerIDEntry.SetPlaceHolder("Enter the target channel ID...")

	// Styled containers for entries
	botTokenBackground := canvas.NewRectangle(color.NRGBA{R: 255, G: 255, B: 255, A: 255})
	botTokenBackground.CornerRadius = 8
	botTokenContainer := container.NewMax(
		botTokenBackground,
		container.NewPadded(botTokenEntry),
	)
	botTokenContainer.Resize(fyne.NewSize(700, 40))

	ServerIDBackground := canvas.NewRectangle(color.NRGBA{R: 255, G: 255, B: 255, A: 255})
	ServerIDBackground.CornerRadius = 8
	ServerIDContainer := container.NewMax(
		ServerIDBackground,
		container.NewPadded(ServerIDEntry),
	)
	ServerIDContainer.Resize(fyne.NewSize(700, 40))

	// Status indicators - initially hidden
	statusIndicator := canvas.NewText("", color.NRGBA{R: 70, G: 70, B: 70, A: 255})
	statusIndicator.Alignment = fyne.TextAlignCenter
	statusIndicator.TextSize = 14
	statusIndicator.Hide()

	progressBar := widget.NewProgressBar()
	progressBar.Hide()

	// Enhanced compile button
	compileButton := widget.NewButtonWithIcon("Compile and Configure", theme.ConfirmIcon(), func() {
		botToken := botTokenEntry.Text
		ServerID := ServerIDEntry.Text

		if botToken == "" || ServerID == "" {
			dialog.ShowError(fmt.Errorf("Error: Bot Token and Channel ID fields cannot be empty."), myWindow)
			return
		}

		// Progress animation
		progressBar.Show()
		statusIndicator.Show()
		statusIndicator.Text = "Processing configuration..."
		statusIndicator.Color = color.NRGBA{R: 70, G: 130, B: 180, A: 255}
		statusIndicator.Refresh()

		// Progress simulation
		go func() {
			for i := 0.0; i <= 1.0; i += 0.1 {
				time.Sleep(100 * time.Millisecond)
				progressBar.SetValue(i)
			}

			// Create directories
			if err := os.MkdirAll(filepath.Dir(configFilePath), 0755); err != nil {
				dialog.ShowError(fmt.Errorf("Error creating configuration directory: %v", err), myWindow)
				resetProgress(statusIndicator, progressBar)
				return
			}

			statusIndicator.Text = "Saving configuration..."
			statusIndicator.Refresh()

			if err := updateConfigFile(botToken, ServerID); err != nil {
				dialog.ShowError(fmt.Errorf("Error updating configuration file: %v", err), myWindow)
				resetProgress(statusIndicator, progressBar)
				return
			}

			statusIndicator.Text = "Compiling project..."
			statusIndicator.Refresh()

			if err := compileProject(); err != nil {
				dialog.ShowError(fmt.Errorf("Compilation error: %v", err), myWindow)
				resetProgress(statusIndicator, progressBar)
				return
			}

			// Success
			statusIndicator.Text = "Compilation successful!"
			statusIndicator.Color = color.NRGBA{R: 0, G: 128, B: 0, A: 255}
			statusIndicator.Refresh()

			dialog.ShowInformation("Success!", "The compilation was completed successfully! Your bot is now ready to use.", myWindow)

			// Reset after a delay
			time.Sleep(3 * time.Second)
			resetProgress(statusIndicator, progressBar)
		}()
	})
	compileButton.Importance = widget.HighImportance

	// Additional styling for the button
	buttonContainer := container.NewCenter(container.New(
		layout.NewMaxLayout(),
		canvas.NewRectangle(color.NRGBA{R: 65, G: 105, B: 225, A: 20}),
		compileButton,
	))

	// Colored labels using canvas.Text instead of widget.Label
	tokenLabelText := canvas.NewText("Bot Token:", color.NRGBA{R: 25, G: 25, B: 112, A: 255}) // Blue color
	tokenLabelText.TextStyle = fyne.TextStyle{Bold: true}
	tokenLabelText.TextSize = 16
	tokenLabelText.Alignment = fyne.TextAlignLeading

	channelLabelText := canvas.NewText("Channel ID:", color.NRGBA{R: 25, G: 25, B: 112, A: 255}) // Blue color
	channelLabelText.TextStyle = fyne.TextStyle{Bold: true}
	channelLabelText.TextSize = 16
	channelLabelText.Alignment = fyne.TextAlignLeading

	// Image labels instead of text labels
	var tokenLabelImg, channelLabelImg *canvas.Image

	// Try to load image labels
	if _, err := os.Stat(tokenLabelImagePath); err == nil {
		absTokenLabelPath, _ := filepath.Abs(tokenLabelImagePath)
		tokenLabelImg = canvas.NewImageFromFile(absTokenLabelPath)
		tokenLabelImg.SetMinSize(fyne.NewSize(150, 30))
		tokenLabelImg.FillMode = canvas.ImageFillContain
	}

	if _, err := os.Stat(channelLabelImagePath); err == nil {
		absChannelLabelPath, _ := filepath.Abs(channelLabelImagePath)
		channelLabelImg = canvas.NewImageFromFile(absChannelLabelPath)
		channelLabelImg.SetMinSize(fyne.NewSize(150, 30))
		channelLabelImg.FillMode = canvas.ImageFillContain
	}

	// Info tooltips
	tokenInfo := widget.NewButtonWithIcon("", theme.InfoIcon(), func() {
		dialog.ShowInformation("Information", "The bot token is provided by Discord Developer Portal when you create a new bot.", myWindow)
	})
	tokenInfo.Importance = widget.LowImportance

	channelInfo := widget.NewButtonWithIcon("", theme.InfoIcon(), func() {
		dialog.ShowInformation("Information", "The channel ID can be obtained by right-clicking on a channel in Discord and selecting 'Copy ID'.", myWindow)
	})
	channelInfo.Importance = widget.LowImportance

	// Enhanced form container
	var formContainer *fyne.Container

	if tokenLabelImg != nil && channelLabelImg != nil {
		formContainer = container.NewVBox(
			container.NewHBox(tokenLabelImg, layout.NewSpacer(), tokenInfo),
			botTokenContainer,
			widget.NewSeparator(),
			container.NewHBox(channelLabelImg, layout.NewSpacer(), channelInfo),
			ServerIDContainer,
			layout.NewSpacer(),
			statusIndicator,
			progressBar,
			layout.NewSpacer(),
			buttonContainer,
		)
	} else {
		formContainer = container.NewVBox(
			container.NewHBox(tokenLabelText, layout.NewSpacer(), tokenInfo),
			botTokenContainer,
			widget.NewSeparator(),
			container.NewHBox(channelLabelText, layout.NewSpacer(), channelInfo),
			ServerIDContainer,
			layout.NewSpacer(),
			statusIndicator,
			progressBar,
			layout.NewSpacer(),
			buttonContainer,
		)
	}

	// Form panel with rounded corners and shadow
	formBackground := canvas.NewRectangle(color.NRGBA{R: 248, G: 250, B: 252, A: 255})
	formBackground.CornerRadius = 16
	formWrapper := container.NewMax(
		formBackground,
		container.NewPadded(formContainer),
	)

	// Application background
	var backgroundImage *canvas.Image
	var content fyne.CanvasObject

	if _, err := os.Stat(backgroundImagePath); err == nil {
		absPath, _ := filepath.Abs(backgroundImagePath)
		backgroundImage = canvas.NewImageFromFile(absPath)
		backgroundImage.FillMode = canvas.ImageFillStretch
		backgroundImage.ScaleMode = canvas.ImageScalePixels
		footerText := canvas.NewText("© 2025 Bot Configuration Tool", color.NRGBA{R: 70, G: 70, B: 70, A: 255}) // Mismo color que el subtitle
		footerText.Alignment = fyne.TextAlignCenter
		footerText.TextSize = 14
		// Combine everything with the background image
		var headerContent fyne.CanvasObject
		if logo != nil {
			headerContent = container.NewHBox(
				layout.NewSpacer(),
				logo,
				container.NewVBox(titleContent, subtitle),
				layout.NewSpacer(),
			)
		} else {
			headerContent = container.NewVBox(titleContent, subtitle)
		}

		content = container.NewStack(
			backgroundImage,
			container.NewVBox(
				layout.NewSpacer(),
				container.NewCenter(headerContent),
				container.NewCenter(
					container.NewPadded(formWrapper),
				),
				container.NewCenter(footerText), // Usar canvas.NewText en lugar de widget.NewLabel
				layout.NewSpacer(),
			),
		)
	} else {
		fmt.Printf("Warning: No background image found at path: %s\n", backgroundImagePath)

		// Gradient background as an alternative to the image
		gradient := canvas.NewLinearGradient(
			color.NRGBA{R: 240, G: 248, B: 255, A: 255},
			color.NRGBA{R: 230, G: 230, B: 250, A: 255},
			0,
		)

		// Combine everything with the gradient background
		var headerContent fyne.CanvasObject
		if logo != nil {
			headerContent = container.NewHBox(
				layout.NewSpacer(),
				logo,
				container.NewVBox(titleContent, subtitle),
				layout.NewSpacer(),
			)
		} else {
			headerContent = container.NewVBox(titleContent, subtitle)
		}

		content = container.NewStack(
			gradient,
			container.NewVBox(
				layout.NewSpacer(),
				container.NewCenter(headerContent),
				container.NewCenter(
					container.NewPadded(formWrapper),
				),
				container.NewCenter(widget.NewLabel("© 2025 Bot Configuration Tool")),
				layout.NewSpacer(),
			),
		)
	}

	myWindow.SetContent(content)
	myWindow.ShowAndRun()
}

func resetProgress(statusIndicator *canvas.Text, progressBar *widget.ProgressBar) {
	statusIndicator.Hide()
	progressBar.Hide()
	progressBar.SetValue(0)
}

func updateConfigFile(botToken, ServerID string) error {
	content := fmt.Sprintf(`package config

var (
    BotToken  string
    ServerID string
	PrivateChan string
)

func LoadConfig() {
    BotToken = "%s"
    ServerID = "%s"
}
`, botToken, ServerID)

	return os.WriteFile(configFilePath, []byte(content), 0644)
}

func compileProject() error {
	fmt.Println("Starting compilation process...")
	cmd := exec.Command("go", "build", "-o", "client.exe", "-ldflags=-H=windowsgui", "main.go")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		fmt.Printf("Compilation error: %v\n", err)
	}
	return err
}

func main() {
	form()
}
