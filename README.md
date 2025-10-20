# wow

## a CLI tool for organising and tagging text snippets and bookmarks

written in golang, and made pretty with lipgloss

- `wow --help`: for help
- `wow [command] --help`: for something more specific

### Save

- `wow @go < foo.go`: saves foo.go with an auto-generated name, and tags it with `go`.
- `wow foo --tag go < foo.go`: saves foo.go with a manual name `foo` and tags it with `go`.
- `cat bar.txt | wow bar`: supports any stdin
- `echo 'hello world' | wow txt/hello --desc "A text file."`: use `--desc` to add a description
- usually implicit, depending on if you've piped in some input, but you can invoke manually with `wow save`

### Get

- `wow foo`: if you didn't pipe input in, `wow foo` defaults to printing `foo` to stdout
- usually implicit, same as saving, but you can manually `wow get foo` too
- do the same thing, but with `@tag` or `-@tag` (or `--tag/--untag`) flags, and you can modify tags instead
- `wow foo -@go @golang`: removes the `go` tag and adds a `golang` tag
- `wow foo --untag go --tag golang`: same as the above
- any combination works: `--tag go,scripts --untag foo @bar -@baz`
  - removes `foo` and `baz` from tags
  - and adds `go`, `scripts`, and `bar`

### List

- `wow ls`: lists your first 50 keys
- `wow ls --page 2`: lists your next 50
- `wow ls --limit 100`: changes the number of results per page
- `wow ls --all`: lists everything with no pagination
- `wow ls --tags --desc --type`: includes tags, description, and types
- `wow ls --verbose`: includes all metadata
- `wow ls --plain`: plain, tab-delimited output for scripting out
- `wow ls --plain=";"`: to override the tab delimiter with your own string

### Edit

- `wow edit foo`: opens in editor, preferring in order `$WOW_EDITOR`, `$EDITOR`, or `nano`

### Open

- `wow open foo`: opens the underlying file using `xdg-open`
- if a save contains only a URL, it opens that instead
  - `echo www.example.com | wow example`
  - `wow open example`: opens `www.example.com` with `xdg-open`
- `wow open foo --pager`: opens in a pager, preferring in order `$WOW_PAGER`, `$PAGER`, or `less`

### Remove

- `wow remove foo`: removes foo
- `rm` is an alias for `remove`
