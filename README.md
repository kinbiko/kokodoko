# ここどこ？　(koko doko?)

[![Build Status](https://github.com/kinbiko/kokodoko/workflows/Go/badge.svg)](https://github.com/kinbiko/kokodoko/actions)
[![Coverage Status](https://coveralls.io/repos/github/kinbiko/kokodoko/badge.svg?branch=main)](https://coveralls.io/github/kinbiko/kokodoko?branch=main)
[![Go Report Card](https://goreportcard.com/badge/github.com/kinbiko/kokodoko)](https://goreportcard.com/report/github.com/kinbiko/kokodoko)
[![Latest version](https://img.shields.io/github/tag/kinbiko/kokodoko.svg?label=latest%20version&style=flat)](https://github.com/kinbiko/kokodoko/releases)
[![Go Documentation](http://img.shields.io/badge/godoc-documentation-blue.svg?style=flat)](https://pkg.go.dev/github.com/kinbiko/kokodoko?tab=doc)
[![License](https://img.shields.io/github/license/kinbiko/kokodoko.svg?style=flat)](https://github.com/kinbiko/kokodoko/blob/master/LICENSE)

Quickly generate GitHub permalink to lines of code in your local filesystem.

## Installation

Assuming you have Go installed:

```
go get -u github.com/kinbiko/kokodoko/cmd/kokodoko
```

## Usage

First argument should be a path to a file/directory on your local system.
Second (optional) argument is the line number, or a line number range, that you would like to highlight.

```console
$ kokodoko ./dotfiles/vimrc 10-15
Copied 'https://github.com/kinbiko/dotfiles/blob/15fc22c0c5672e0f15f2ef7ea333bd620aa9965c/vimrc#L10-L15' to the clipboard!
```

However, this tool is more useful when integrated directly in your editor.

### Vim integration

Put the following in your `.vimrc`:

```vim
" Super hacky mapping to shell out to github.com/kinbiko/kokodoko and fetch the
" Github link to the current line(s)
nnoremap <silent> oiuy :!kokodoko % <C-R>=line(".")<CR><CR>
vnoremap <silent> oiuy :!kokodoko % <C-R>=line("'<")<CR>-<C-R>=line("'>")<CR><CR>u
```

Replace `oiuy` with your desired mapping.

### Integrating with other editors

Contributions are welcome!
