; Big Top Builder Classic (Demo) - standalone setup. Build: bash build.sh, then ISCC.exe /DGameVersion=<version> installer\BigTopSetup.iss
; (c) 2026 Ironite
#define GameName    "Big Top Builder Classic"
#define OldGameName "Big Top Builder"
#define Bundle      "BigTopBuilder"
#define Exe         "BigTopBuilder.exe"
#ifndef GameVersion
  #define GameVersion "0.1.0"
#endif
#define AppGuid     "{{4F7A2C93-1D6B-4E58-A3F0-9B2E7C51D864}"
#define Art         "bigtop"
#define Tagline     "A circus management sim (free demo). Raise the tents, hire the troupe, plan the shows and keep the crowd cheering."
#define SaveWhat    "Big Top Builder Classic settings and records"
#include "GameSetup.iss"
