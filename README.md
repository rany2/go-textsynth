# go-textsynth

Text Synth client in Go.

## To build

* `git clone https://github.com/rany2/go-textsynth.git`
* `go build ./cmd/go-textsynth`
* (or you can use `go install ...`, whichever you prefer)

## Usage

```
$ go-textsynth -h 
Usage of go-textsynth:
  -apikey string
    	API key for Text Synth (default "842a11464f81fc8be43ac76fb36426d2")
  -dont-normalize-newline
    	Do not convert Windows and Mac OS line endings to Unix
  -max-tokens uint
    	Maximum number of tokens to generate. (default 200)
  -model string
    	Select a model (boris_6B, fairseq_gpt_13B, and gptj_6B) (default "gptj_6B")
  -prompt string
    	Prompt to send to Text Synth
  -promptfile string
    	Like prompt but read from file
  -seed uint
    	Seed of the random number generator. Use 0 for a random seed.
  -stop string
    	Stop token to stop generation. Use "" to disable.
  -temperature float
    	Divide the logits (=log(probability) of the tokens) by the temperature value (0.1 <= temperature <= 10) (default 1)
  -top-k float
    	Keep only the top-k tokens with the highest probability (1 <= top-k <= 1000) (default 40)
  -top-p float
    	Keep the top tokens having cumulative probability >= top-p (0 < top-p <= 1) (default 0.9)
```

## Notes

If the results aren't what you wanted, you might want to try to
[set `-top-p` to `1.0` and `-temperature` to something between
`0.7` and `0.8`](https://news.ycombinator.com/item?id=27727257).
The results do seem better when using those settings but I'll keep
using the Text Synth website's defaults for this program.

In order to use `go-textsynth`, you must provide a prompt via
either `-prompt` or `-promptfile`. `-prompt` takes the argument
after it as the prompt while `-promptfile` takes the argument
after it as the file path to read the prompt from.

When using `-promptfile` you probably want a text editor that
doesn't automatically add newlines to the end of the file.
By default, CR or CRLF line endings are replaced with LF; you
could use `-dont-normalize-newline` to disable this. However,
`go-textsynth` doesn't do anything about trailing newlines,
trailing spaces, repeating spaces, etc.
