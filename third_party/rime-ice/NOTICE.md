# Third-party: rime-ice

- Project: https://github.com/iDvel/rime-ice
- Local cache: `8105.dict.yaml`, `41448.dict.yaml`, `base.dict.yaml`, `ext.dict.yaml`, `tencent.dict.yaml`, `others.dict.yaml`, `opencc/emoji.txt`, `opencc/others.txt`, `opencc/emoji.json`, `LICENSE`
- Upstream license: GPL-3.0 (see `LICENSE`)
- Usage in TSInputMethod: offline source for `tools/dictBuilder` and shipped emoji map
- Runtime: compiled `.tsdict` index + `emoji.txt` loaded by Go; **no Rime Runtime**
- Redistribution: keep LICENSE and source YAML available with derived indexes
