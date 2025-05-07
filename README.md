# Discord Rich Presence for Zathura document viewer

![Example rich presence](https://github.com/user-attachments/assets/75d0fbfb-53b9-4672-8179-7e993ca6b908)

Application for updating your Discord activity with metadata from opened documents in Zathura.

## How to use

1. Ensure Zathura is installed and working
2. Download the source code
3. Ensure Golang version >= 1.24.3 is installed
6. Building the source code with `go build` and run the executable (e.g: `./ZathuraDiscordRichPresence` after building on Linux)
7. Start Zathura (documents openend in Zathura with sandbox mode are NOT detected by this application)
8. (Optional) open a document in Zathura
8. Your Discord rich presence should now be updated

By default the hover presence hover data shows a generic message:

![250507231139](https://github.com/user-attachments/assets/5f349000-ddf1-4bfa-8050-926a3cbe9b65)

This message can optionally be changed to show the current chapter that is being read, by running the application with the `-show-chapters` flag
- `./ZathuraDiscordRichPresence`.go -show-chapters`:

![250507231114](https://github.com/user-attachments/assets/a6fe91ef-5b37-424b-944a-35e8ab254d02)

HOWEVER, be weary that this data could include sensitive information depending on which document you open (which is why it is turned off by default.

The filename shown in the precense is taken directly from the document's filename (stripped from the full path returned by the Zathura d-bus interface). Ensure that opened documents while the extension is running contain no sensitive filenames.

## How it works

- This application continously polls the d-bus interface implemented in Zathura to retrieve current opened document information.

## Limitations and potential improvements
- Ideally Zathura would emit events whenever a document is opened, a document is closed, and when page switches happen. This way applications like this do not have to continously poll to see whether anything has changed, but can instead listen to the d-bus event queue (observer). I created a ![pull-request](https://github.com/pwmt/zathura/pull/742) on the Zathura repository that implements DocumentOpen and DocumentClose events, so if these ever get added I will look forward to improving this application.
- Only tested on GNU/Linux

## Attribution
- ![Rich-go for the Discord Rich Presence implementation](https://github.com/hugolgst/rich-go)
- ![Godbus for the d-bus implementation](https://github.com/godbus/dbus)
- Assets (document icons in the rich presence)
  - <a href="https://www.flaticon.com/free-icons/format" title="format icons">Format icons created by JunGSa - Flaticon</a>
  - <a href="https://www.flaticon.com/free-icons/pdf" title="pdf icons">Pdf icons created by JunGSa - Flaticon</a> (colored red by me :p)
  - <a href="https://www.flaticon.com/free-icons/epub" title="epub icons">Epub icons created by JunGSa - Flaticon</a>
