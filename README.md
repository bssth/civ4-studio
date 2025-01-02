# Civilization 4 Map Studio

Welcome to the **Civilization 4 Map Studio**! This tool is designed to help you in editing and creating custom maps for [Civilization 4](https://en.wikipedia.org/wiki/Civilization_IV), turn-based strategy game. 

While the WordBuilder-like interface is not implemented for now, this editor provides functionality beyond the capabilities of the standard WorldBuilder, allowing you to tweak various aspects of your maps.

## Features

1. Edit Civilization and Leader Lists: Customize the civilizations and leaders available in your game.
2. Fine-tune Map Settings: Refine map settings including size, shape, and starting positions.
3. Advanced Options: Explore additional parameters not accessible in the standard editor.

## Getting Started

Download release from GitHub and run!

## Building from sources

Clone or download the repository to your local machine.

You now need to download the Fyne module and helper tool. This will be done using the following commands:

```bash
$ go get fyne.io/fyne/v2@latest
$ go install fyne.io/fyne/v2/cmd/fyne@latest
$ go install github.com/fyne-io/fyne-cross@latest
$ fyne-cross windows # (or fyne-cross darwin, fyne-cross linux etc.)
```

## Contributing

We welcome contributions to this project. If you would like to contribute, please fork the repository and submit a pull request. We will review your changes and merge them into the main branch if they are deemed appropriate.

Please follow [Golang styleguide](https://google.github.io/styleguide/go/) and use `gofmt` to format your code.

Using Goland is strongly recommended, but if you are not familiar with it, you can use any other IDE you are comfortable with.

### Cautions and known problems

1. The first run from source code can take a very long time, as Fyne and its dependencies need to be compiled. Be patient, it may take a few minutes.
2. The interface code is implemented to be clear and intuitive, but can be intimidating with the amount of boilerplate. Please refrain from making fundamental changes without discussion.
3. WebAssembly is supported thanks to Fyne, but is not yet adapted and tested. If you need it for any reason, please open an issue.
4. You may encounter "@todo" markings in the code. You can implement and contribute what is marked, unless otherwise explicitly stated in the comment.
5. I love French hot dogs 😋

## Support

For any questions, issues, or suggestions, please feel free to contact us using Issues. Your feedback is valuable in improving this tool for the Civilization 4 community.

This editor is a work in progress and may contain bugs or incomplete features. Use it at your own risk. We are continually working to enhance its functionality and reliability.

**Please remember to back up your maps before editing them.**

Happy mapping!

