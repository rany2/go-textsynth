Text Synth client in Go.

To build:

* `git clone https://github.com/rany2/go-textsynth.git`
* `go build ./cmd/go-textsynth`
* (or you can use `go install ...`, whichever you prefer)

If the results aren't what you wanted, you might want to try to
[set `-top-p` to `1.0` and `-temperature` to something between
`0.7` and `0.8`](https://news.ycombinator.com/item?id=27727257).
The results do seem better when using those settings but I'll keep
using the Text Synth website's defaults for this program.

Note that when using `-promptfile` you probably want a text editor
that doesn't automatically add newlines to the end of the file.
The data you send to `go-textsynth` is not modified in any way
(no trimming, the entire prompt file is sent, etc).