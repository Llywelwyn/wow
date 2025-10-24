# pda

## a CLI tool for organising and tagging text snippets and bookmarks

written in golang, and made pretty with lipgloss

- `pda --help`: for help
- `pda [command] --help`: for something more specific

### Save

- `pda @go < foo.go`: saves foo.go with an auto-generated name, and tags it with `go`.
- `pda foo --tag go < foo.go`: saves foo.go with a manual name `foo` and tags it with `go`.
- `cat bar.txt | pda bar`: supports any stdin
- `echo 'hello world' | pda txt/hello --desc "A text file."`: use `--desc` to add a description
- usually implicit, depending on if you've piped in some input, but you can invoke manually with `pda save`

### Get

- `pda foo`: if you didn't pipe input in, `pda foo` defaults to printing `foo` to stdout
- usually implicit, same as saving, but you can manually `pda get foo` too
- do the same thing, but with `@tag` or `-@tag` (or `--tag/--untag`) flags, and you can modify tags instead
- `pda foo -@go @golang`: removes the `go` tag and adds a `golang` tag
- `pda foo --untag go --tag golang`: same as the above
- any combination works: `--tag go,scripts --untag foo @bar -@baz`
  - removes `foo` and `baz` from tags
  - and adds `go`, `scripts`, and `bar`

### List

- `pda ls`: lists your first 50 keys
- `pda ls --page 2`: lists your next 50
- `pda ls --limit 100`: changes the number of results per page
- `pda ls --all`: lists everything with no pagination
- `pda ls --tags --desc --type`: includes tags, description, and types
- `pda ls --verbose`: includes all metadata
- `pda ls --plain`: plain, tab-delimited output for scripting out
- `pda ls --plain=";"`: to override the tab delimiter with your own string

### Edit

- `pda edit foo`: opens in editor, preferring in order `$pda_EDITOR`, `$EDITOR`, or `nano`

### Open

- `pda open foo`: opens the underlying file using `xdg-open`
- if a save contains only a URL, it opens that instead
  - `echo www.example.com | pda example`
  - `pda open example`: opens `www.example.com` with `xdg-open`
- `pda open foo --pager`: opens in a pager, preferring in order `$pda_PAGER`, `$PAGER`, or `less`

### Remove

- `pda remove foo`: removes foo
- `rm` is an alias for `remove`
