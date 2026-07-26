# Atribución — `precios.json` deriva de LiteLLM (MIT)

`precios.json` se genera con `go generate ./internal/adapters/telemetria/catalogo/`, que baja
`model_prices_and_context_window.json` de un **rev fijado** del repo
[BerriAI/litellm](https://github.com/BerriAI/litellm), lo filtra a los proveedores que este árbol
puede ver y lo compacta.

- **Rev fijado:** `b439a9a78864f52fba3d18d9bf8b09253e7181d7` (2026-07-26).
- **Archivo de origen:** `model_prices_and_context_window.json`, en la **raíz** del repo. El
  carve-out comercial de LiteLLM vive bajo `enterprise/` y **no alcanza a este archivo**.
- **Licencia del origen:** MIT.

## Licencia MIT de LiteLLM (texto)

```
MIT License

Copyright (c) 2023 Berri AI

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
```

## Qué se cambió respecto del origen, y por qué

1. **Se filtró por proveedor** (anthropic · bedrock · vertex · openai · gemini · deepseek · xai ·
   mistral · groq) y por `mode ∈ {chat, responses}`: 2 984 modelos → 694. El resto no cotiza
   tokens de un runtime de agente y solo pesaría en el binario.
2. **Se descartaron los campos que no cotizan tokens** (`supports_*`, `max_tokens`,
   `search_context_cost_per_query`…).
3. **NO se aplanaron los tiers.** `cache_creation_input_token_cost_above_1hr` y el tramo
   `above_200k` viajan enteros. Es la diferencia con ccusage, que colapsa todo a un solo
   `cache_create` y pierde justo el campo que habilita el detector B1 (re-warm por TTL).
