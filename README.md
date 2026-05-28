# ssh-portfolio

zena's personal website that lives in the terminal.

```
ssh zena-portfolio.fly.dev
```

## stack

- [Wish](https://github.com/charmbracelet/wish) — SSH server
- [Bubble Tea](https://github.com/charmbracelet/bubbletea) — TUI framework
- [Lip Gloss](https://github.com/charmbracelet/lipgloss) — terminal styling
- [Fly.io](https://fly.io) — hosting

## features

- braille-art portrait (preprocessed photo → `portrait.txt`, embedded at build time)
- falling-snow animation overlaid on the figlet name
- per-tab color palette (cyan / mint / purple / magenta / pink)
- rotating fortune line
- hidden `?` panel with a short reflection on AI, creativity, and human purpose
- works in any modern terminal that supports truecolor

## run locally

```bash
# 1. generate a host key (one time, ignored by git)
mkdir -p .ssh
ssh-keygen -t ed25519 -f .ssh/host_key -N ""

# 2. install dependencies
go mod tidy

# 3. run — defaults to port 2222 locally so no sudo is needed
go run .

# 4. connect in another terminal
ssh localhost -p 2222
```

Press `?` for the hidden reflection. `← →` to switch tabs. `f` to roll a new fortune. `q` to quit.

## regenerating the portrait

The braille art in `portrait.txt` was generated from a source photo with:

```bash
# crop + sharpen the source photo (ImageMagick)
magick me.jpeg -crop 550x900+180+50 +repage \
  -unsharp 0x3+1.5+0 -modulate 100,115 /tmp/me.jpg

# render to braille (ascii-image-converter)
ascii-image-converter --braille -d 50,36 /tmp/me.jpg > portrait.txt
```

Tools:

```bash
brew install imagemagick ascii-image-converter
```

## deploy to fly.io

```bash
# 1. install flyctl + log in
brew install flyctl
fly auth login

# 2. generate a persistent host key (ignored by git)
ssh-keygen -t ed25519 -f .ssh/prod_host_key -N ""

# 3. create the app + upload the host key as a secret
fly apps create zena-portfolio
fly secrets set HOST_KEY_B64="$(base64 < .ssh/prod_host_key)" --app zena-portfolio

# 4. deploy
fly deploy --app zena-portfolio
```

The Dockerfile is a multi-stage build that produces a static binary in a `scratch` image. At runtime the app:

- reads `HOST_KEY_B64` from env, decodes it, and writes it to `HOST_KEY_PATH` so the SSH fingerprint stays stable across deploys
- listens on `PORT` (Fly maps external `22` → internal `2222`, since Fly's own `hallpass` SSH server already binds port `22` inside every container)

## custom domain

```bash
fly certs add ssh.example.com --app zena-portfolio
```

Add an `A` and `AAAA` DNS record to the IPs from `fly ips list`.

> If you're behind Cloudflare, set the DNS record to **DNS-only** (grey cloud). SSH is raw TCP and Cloudflare's HTTP proxy will break it.

## license

MIT
