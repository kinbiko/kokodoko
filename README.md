# kokodoko

Quickly generate GitHub permalink to lines of code in your local filesystem.

## Installation

Assuming you have Go installed:

```
go get -u github.com/kinbiko/kokodoko
```

## Usage

First argument should be a path to a file/directory on your local system.
Second (optional) argument is the line number, or a line number range, that you would like to highlight.

```console
$ kokodoko ./dotfiles/vimrc 10-15
Copied 'https://github.com/kinbiko/dotfiles/blob/15fc22c0c5672e0f15f2ef7ea333bd620aa9965c/vimrc#L10-L15' to the clipboard!
```
