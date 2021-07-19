# go-textsynth

Text Synth client in Go.

## To build

* `git clone https://github.com/rany2/go-textsynth.git`
* `go build ./cmd/go-textsynth`
* (or you can use `go install ...`, whichever you prefer)

## Notes

If the results aren't what you wanted, you might want to try to
[set `-top-p` to `1.0` and `-temperature` to something between
`0.7` and `0.8`](https://news.ycombinator.com/item?id=27727257).
The results do seem better when using those settings but I'll keep
using the Text Synth website's defaults for this program.

When using `-promptfile` you probably want a text editor that
doesn't automatically add newlines to the end of the file.
By default, CR or CRLF line endings are replaced with LF; you
could use `-dont-normalize-newline` to disable this. However,
`go-textsynth` doesn't do anything about trailing newlines,
trailing spaces, repeating spaces, etc.
