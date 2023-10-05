# ctoc

_Count Tokens of Code_.

> Token counts plays a key role in shaping an LLM's memory and conversation history.<br/>
> **ctoc** provides a lightweight tool to analyze codebases at the token level.
>
> Built on top of [gocloc](https://github.com/hhatto/gocloc).

[![GoDoc](https://godoc.org/github.com/yaohui-wyh/ctoc?status.svg)](https://godoc.org/github.com/yaohui-wyh/ctoc)
[![ci](https://github.com/yaohui-wyh/ctoc/workflows/Go/badge.svg)](https://github.com/yaohui-wyh/ctoc/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/hhatto/gocloc)](https://goreportcard.com/report/github.com/yaohui-wyh/ctoc)

<details>
<summary>What are <b>Tokens</b>? (in the context of Large Language Model)</summary> 

> https://learn.microsoft.com/en-us/semantic-kernel/prompt-engineering/tokens

- **Tokens**: basic units of text/code for LLM AI models to process/generate language.
- **Tokenization**: splitting input/output texts into smaller units for LLM AI models.
- **Vocabulary size**: the number of tokens each model uses, which varies among different GPT models.
- **Tokenization cost**: affects the memory and computational resources that a model needs, which influences the cost
  and performance of running an OpenAI or Azure OpenAI model.
</details>

## Installation

require Go 1.19+

```
$ go install github.com/yaohui-wyh/ctoc/cmd/ctoc@latest
```

## Usage

### Basic Usage

```
$ ctoc -h
Usage:
  ctoc [OPTIONS]

Application Options:
      --by-file                                              report results for every encountered source file
      --sort=[name|files|blank|comment|code|tokens]          sort based on a certain column (default: code)
      --output-type=                                         output type [values: default,cloc-xml,sloccount,json] (default: default)
      --exclude-ext=                                         exclude file name extensions (separated commas)
      --include-lang=                                        include language name (separated commas)
      --match=                                               include file name (regex)
      --not-match=                                           exclude file name (regex)
      --match-d=                                             include dir name (regex)
      --not-match-d=                                         exclude dir name (regex)
      --debug                                                dump debug log for developer
      --skip-duplicated                                      skip duplicated files
      --show-lang                                            print about all languages and extensions
      --version                                              print version info
      --encoding=[cl100k_base|p50k_base|p50k_edit|r50k_base] specify tokenizer encoding (default: cl100k_base)

Help Options:
  -h, --help                                                 Show this help message
```

```
$ ctoc .
------------------------------------------------------------------------------------------------
Language                     files          blank        comment           code           tokens
------------------------------------------------------------------------------------------------
Go                              15            282            153           2096          21839
XML                              3              0              0            140           1950
YAML                             1              0              0             40            237
Markdown                         1             13              0             34            322
Makefile                         1              6              0             15            128
------------------------------------------------------------------------------------------------
TOTAL                           21            301            153           2325          24476
------------------------------------------------------------------------------------------------
```

## Support Languages

> Same as [gocloc](https://github.com/hhatto/gocloc#support-languages)

```
$ ctoc --show-lang
```

## Support Models

```
$ ctoc --show-encoding
text-davinci-002               (p50k_base)
text-davinci-001               (r50k_base)
babbage                        (r50k_base)
text-babbage-001               (r50k_base)
code-cushman-002               (p50k_base)
code-search-ada-code-001       (r50k_base)
text-davinci-003               (p50k_base)
davinci                        (r50k_base)
text-similarity-ada-001        (r50k_base)
text-curie-001                 (r50k_base)
curie                          (r50k_base)
ada                            (r50k_base)
code-davinci-002               (p50k_base)
text-davinci-edit-001          (p50k_edit)
text-embedding-ada-002         (cl100k_base)
text-similarity-curie-001      (r50k_base)
text-similarity-babbage-001    (r50k_base)
gpt2                           (gpt2)
gpt-4                          (cl100k_base)
text-ada-001                   (r50k_base)
code-davinci-001               (p50k_base)
text-search-davinci-doc-001    (r50k_base)
text-search-curie-doc-001      (r50k_base)
code-search-babbage-code-001   (r50k_base)
code-cushman-001               (p50k_base)
cushman-codex                  (p50k_base)
code-davinci-edit-001          (p50k_edit)
gpt-3.5-turbo                  (cl100k_base)
text-similarity-davinci-001    (r50k_base)
text-search-babbage-doc-001    (r50k_base)
text-search-ada-doc-001        (r50k_base)
davinci-codex                  (p50k_base)
```

The BPE dictionary is automatically downloaded and cached upon its initial run for each encoding.<br/>
For additional information, please refer to [tiktoken-go#cache](https://github.com/pkoukk/tiktoken-go#cache)

## Performance

- CPU 2.6GHz 6core Intel Core i7 / 32GB 2667MHz DDR4 / MacOSX 13.5.2
- ctoc [fdaa42](https://github.com/yaohui-wyh/ctoc/commit/fdaa42)

```
➜  kubernetes git:(master) time ctoc .
------------------------------------------------------------------------------------------------
Language                     files          blank        comment           code           tokens
------------------------------------------------------------------------------------------------
Go                           15172         503395         992193        3921496       53747627
JSON                           430              2              0        1011821       10428573
YAML                          1224            612           1464         156024         974131
Markdown                       461          24842            170          93141        3251948
BASH                           318           6522          12788          33010         528217
Protocol Buffers               130           5864          19379          12809         358110
Assembly                        50           2212            925           8447         129534
Plain Text                      31            203              0           6664          48218
Makefile                        58            594            940           2027          31548
Bourne Shell                     9            154            119            687           8055
sed                              4              4             32            439           3138
Python                           7            114            160            418           5435
Zsh                              1             14              3            191           1872
PowerShell                       3             44             79            181           2496
C                                5             42             55            140           1799
TOML                             6             31            107            101           2049
HTML                             2              0              0              2             21
Batch                            1              2             17              2            170
------------------------------------------------------------------------------------------------
TOTAL                        17912         544651        1028431        5247600       69522941
------------------------------------------------------------------------------------------------
ctoc .  160.09s user 8.08s system 119% cpu 2:20.96 total`
```


## License

MIT
