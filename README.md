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

Clone or download the repository to your local machine. You need:

- Go 1.22+
- Node.js 18+ and npm
- The [Wails v2 CLI](https://wails.io/docs/gettingstarted/installation):
  `go install github.com/wailsapp/wails/v2/cmd/wails@latest`

Run the editor in development mode (hot-reload for both frontend and backend):

```bash
$ wails dev
```

Build a production binary for your current platform:

```bash
$ wails build
```

The output binary appears under `build/bin/`. To target another OS, see
[`wails build --help`](https://wails.io/docs/reference/cli) (cross-compilation
of the Windows webview backend has its own requirements).

## Contributing

We welcome contributions to this project. If you would like to contribute, please fork the repository and submit a pull request. We will review your changes and merge them into the main branch if they are deemed appropriate.

Please follow [Golang styleguide](https://google.github.io/styleguide/go/) and use `gofmt` to format your code.

Using Goland is strongly recommended, but if you are not familiar with it, you can use any other IDE you are comfortable with.

### Cautions and known problems

1. The first `wails dev` / `wails build` may take a while as Go and npm dependencies are downloaded and compiled.
2. On Windows, launching the game from the editor uses `ShellExecute` with the `runas` verb and will trigger a UAC prompt.
3. You may encounter "@todo" markings in the code. You can implement and contribute what is marked, unless otherwise explicitly stated in the comment.
4. I love French hot dogs 😋

## Support

For any questions, issues, or suggestions, please feel free to contact us using Issues. Your feedback is valuable in improving this tool for the Civilization 4 community.

This editor is a work in progress and may contain bugs or incomplete features. Use it at your own risk. We are continually working to enhance its functionality and reliability.

**Please remember to back up your maps before editing them.**

Happy mapping!

