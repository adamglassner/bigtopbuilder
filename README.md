# Big Top Builder (Demo)

A circus management sim by Ironite, written in Go with [Ebitengine](https://ebitengine.org).
Raise the tents, hire the troupe, plan each show's running order and keep the crowd cheering from 6 to
11 PM. This is the free demo: run your circus in Brookhaven for as many nights as you like. It ships on its own and inside the
[Ironite Launcher](https://github.com/adamglassner/launcher), where it's listed under Simulation as a DEMO.

## How to play

- **Daytime:** pick something from the bar at the bottom and click the grounds to build it. Guests line
  up at the gold square in front. Click anything you've built to set its price or sell it. Pick **Path**
  from Extras and click and drag to lay paths; guests prefer them, walk faster on them and stay out of the
  mud. Drag over a path to pick it up.
- **The troupe:** open **Troupe** to hire clowns, stilt walkers, jugglers, strongwomen, acrobats, fire
  breathers and trapeze artists. You start with a clown and a juggler; new faces turn up every morning.
- **Plan the shows:** click a tent and book acts into its lineup (three in a Ring Tent, five in the Big
  Top), then use the arrows to set the running order. Shows are rated out of five stars: open with Laughs,
  don't put two acts of the same kind back to back, and close with Thrills. The trapeze needs the Big Top.
- **Showtime:** press **Open the Gates**. Tents wait for an audience, then put on the show act by act.
  Every show tires the acts out and tired acts can flub, so swap them around. Space pauses; 1, 2 and 3 set
  the speed. Nights go on as long as you like; the run ends only if you go broke.
- Esc cancels or opens the menu, M toggles music, F11 toggles full screen, Delete sells what's selected.

Settings, records and achievements are kept in `%APPDATA%\Ironite\BigTopBuilder\save.json`, and the
achievements Ironite Launcher shows in `%APPDATA%\Ironite\Achievements\bigtop.json`.

## Play

The easiest way is **[Ironite Launcher](https://github.com/adamglassner/launcher)**: it downloads the game from
this repository's releases, installs it and keeps it up to date.

Or download from the [latest release](../../releases/latest): `BigTopBuilder-Setup-v<version>.exe`
installs it with Start menu and desktop shortcuts, or unzip `BigTopBuilder-<version>-windows.zip`
anywhere and run `BigTopBuilder.exe`.

## Code

| File | What it is |
|---|---|
| `internal/circus/` | The simulation: the grounds, tents and their shows, performers, guests, money and nights (no graphics; has tests). |
| `internal/synth/` | Synthesizes the circus march and the sound effects at start-up (has tests). |
| `main.go`, `game.go` | Window setup, input, the game loop and screenshot hooks. |
| `art.go` | The tents, sideshows, food carts and performers. |
| `scene.go` | Draws the grounds, guests, night lighting and spotlights. |
| `ui.go`, `troupe.go`, `screens.go` | The top bar, side panel and build bar; the troupe and lineup editor; title, menu, help, settings and end-of-night screens. |
| `draw.go` | Vector drawing and text on a supersampled canvas. |
| `sound.go`, `save.go` | Audio playback; the save file and achievements. |
| `tools/art/` | The icon (SVG) and the script that renders it for the game, the .exe and the installer. |
| `tools/mkres/` | Writes the Windows icon and version info into the .exe. |

## Building

```bash
bash build.sh          # needs Go 1.24+; output in dist/BigTopBuilder/
go test ./internal/... # simulation and audio tests
```

## Releasing

Add a new section at the top of `CHANGELOG.md` with the new version and push to `main`. GitHub Actions tests,
builds and publishes `v<version>` with the zip the launcher installs and a standalone setup.

© 2026 Ironite. All rights reserved. Typefaces: Rye, Oswald and Rubik (SIL Open Font License, see
`assets/fonts/OFL.txt`).
