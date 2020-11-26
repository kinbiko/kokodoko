# ここどこ？　(koko doko?)

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
" Super hacky plugin to shell out to github.com/kinbiko/kokodoko and fetch the
" Github link to the current line(s)
nnoremap <silent> oiuy :!kokodoko % <C-R>=line(".")<CR><CR>
vnoremap <silent> oiuy :!kokodoko % <C-R>=line("'<")<CR>-<C-R>=line("'>")<CR><CR>u
```

Replace `oiuy` with your desired mapping.

### Integrating with other editors

Contributions are welcome!
