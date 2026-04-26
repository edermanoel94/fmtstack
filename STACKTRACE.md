# Reading a Go stack trace

A quick reference for the parts of a Go stack trace — the kind of input `fmtstack` colorizes. Use it when `{_, _}` or `+0x2c` leave you guessing.

## Shape of a trace

Each goroutine contributes a block: a header followed by frames in pairs of **(function line, file:line)**, top of the stack first.

```
goroutine 67 [running]:                                ← goroutine header
runtime/debug.Stack()                                  ┐
        /usr/local/go/src/runtime/debug/stack.go:26 +0x5e   ┘ frame 1 (top)
github.com/sourcegraph/conc/panics.NewRecovered(...)   ┐
        .../panics.go:59 +0x85                              ┘ frame 2
...
created by github.com/sourcegraph/conc.(*WaitGroup).Go in goroutine 8
        .../waitgroup.go:30 +0x73                      ← where this goroutine was spawned
```

The **top** frame is where the panic (or `debug.Stack`) happened. The **bottom** frame shows how the goroutine started.

## Goroutine header

```
goroutine 67 [running]:
│         │  └── current state
│         └──── goroutine ID
└────────────── start of a block
```

The state in brackets comes from one of two tables in the runtime:

- **G status** (scheduler's view): `idle`, `runnable`, `running`, `syscall`, `waiting`, `dead`, `copystack`, `preempted`, `waiting for cgo callback`.
- **Wait reason** (swapped in when the G status is `waiting`, for a more useful label): `chan send`, `chan receive`, `select`, `IO wait`, `sync.Mutex.Lock`, `semacquire`, `sleep`, `GC assist wait`, and many more.

If a goroutine has been blocked for a while, the runtime appends `, N minutes` — useful for spotting leaks.

## Function line

```
findit/internal/clients/googlemaps . (*Client) . Reverse ( 0xc00011e1c0 , {0xc0000f5136, 0x9} , {0xc0000f5140, 0x9} )
└────────── 1 ──────────────────┘    └── 2 ──┘   └─ 3 ─┘   └──── 4 ────┘   └──────── 5 ───────┘   └──────── 6 ───────┘
```

1. package path
2. receiver type — `*` means **pointer receiver**
3. method name
4. receiver value (pointer to a `Client`)
5. 1st argument — a `string`, shown as `{ptr, len=9}`
6. 2nd argument — another `string`

Package-level functions skip parts 2 and 3.

### How each type prints

Each Go type takes a fixed number of **words** (8 bytes on 64-bit). The runtime prints those words in memory order:

| Type                        | Words | Looks like                        |
|-----------------------------|-------|------------------------------------|
| `*T`, `chan`, `map`, `func` | 1     | `0xc00011e1c0`                     |
| `int`, `bool`, `uintptr`    | 1     | `0x1`                              |
| `string`                    | 2     | `{ptr, len}`                       |
| `[]T`                       | 3     | `{ptr, len, cap}`                  |
| `interface{}` / `any`       | 2     | `{type_ptr, value_ptr}`            |
| `struct{A; B; C}`           | sum   | `{<A>, <B>, <C>}` recursively      |

So `{0xb37c60, 0xc000238890}` in an `any` argument reads as **"value of type `0xb37c60`, data at `0xc000238890`"** — not two unrelated numbers.

### `_` and `?`

```
findit/internal/services.(*LocationStrategy).Execute(_, {{_, _}, {_, _}, {_, _}, {_, _}})
                                                     │   │
                                                     │   └── struct of 4 fields, each 2 words
                                                     └────── argument offset can't be encoded
```

- **`_`** — the compiler **couldn't encode** this argument's location in the function's metadata table (`abi.TraceArgsOffsetTooLarge` in `runtime/traceback.go`). Happens with very large structs or deep nesting where the offset doesn't fit in the encoding. You see the **shape** of the argument but no bytes.
- **`<value>?`** — the runtime printed what it found, but **liveness analysis at this PC can't confirm** the slot still holds the original argument. Common with the register-based ABI where registers get reused. Treat it as a hint, not a fact.

```
github.com/sourcegraph/conc/pool.(*Pool).worker(0xc000343d28?)
                                                            │
                                                            └── liveness can't confirm this is the original
```

> Two different problems: `_` is about the **compiler's metadata** (the offset doesn't fit); `?` is about **runtime liveness** (the bytes might be stale). The `?` marker was added in Go 1.18 to make register-passed arguments honest after the register-based calling convention landed in 1.17.

## File:line line

```
        /home/eder/Workspace/findit-rio/internal/clients/googlemaps/client.go : 55 +0x5c5
        └─────────────────────────── 1 ──────────────────────────────────────┘   │   └─ 3 ─┘
                                                                                 └─ 2
```

1. absolute source file path
2. line number
3. PC offset in hex bytes from the start of the function's compiled code

### What `+0x5c5` is for

It's the **program counter offset** inside the compiled function. Why you might care:

- **Disambiguates calls on the same source line** — `f(g(), h())` reports the same `:55` twice with different offsets.
- **Pinpoints instructions** for `go tool addr2line` or `objdump`.
- **Distinguishes compiler-generated wrappers** (closures, `defer`, interface dispatch) that share a `file:line`.

Day-to-day you rarely look at it — that's why `fmtstack` paints it gray.

## `created by ...` frame

```
created by github.com/sourcegraph/conc.(*WaitGroup).Go in goroutine 8
│          │                                              │
│          │                                              └── parent goroutine ID
│          └────────────────────────────────────────────────── function that ran the `go`
└─────────────────────────────────────────────────────────────  marks where this goroutine came from
        /home/eder/.../waitgroup.go:30 +0x73     ← line in the parent's source with the `go func() {...}`
```

Always the **last frame** in a block. When a library (`conc`, `errgroup`) does the `go`, this points at the library, not your code — to find your call site, scroll up to `goroutine 8` in the same trace.

## Knobs that change what you see

- **`GOTRACEBACK`** (env var) controls verbosity:
  - `none` — no traceback at all
  - `single` (default) — current goroutine, user frames only
  - `all` — all goroutines, user frames only
  - `system` — all goroutines including runtime frames
  - `crash` — like `system`, then crash with an OS-level core dump
- **`GODEBUG=tracebacklabels=1`** (Go 1.26+) includes `runtime/pprof` labels in the goroutine status header — the key/value pairs appear next to the state.

## Glossary

| Term | Meaning |
|------|---------|
| **PC** (program counter) | CPU register holding the address of the current machine instruction. The runtime captures it per frame to derive both `file:line` and the `+0xNN` offset. |
| **frame** | One entry on the stack — a function call in progress, with its arguments and locals. |
| **word** | Native machine word: 8 bytes on 64-bit, 4 on 32-bit. The runtime prints arguments one word at a time. |
| **slot** | A storage location for one word — either a CPU register or a fixed offset on the stack. |
| **ABI** (application binary interface) | Rules for how compiled functions pass arguments and return values. Go switched to a **register-based ABI** in 1.17 — small args travel in registers instead of on the stack, which is why `?` markers exist. |
| **liveness analysis** | Compiler/runtime check that says whether a slot still holds a particular value at a given PC. When it can't confirm, the runtime prints `?`. |
| **inlining** | Compiler optimization that pastes a function body into its caller. Inlined functions don't get their own frame. |
| **G** | Internal name for a goroutine (`runtime.g`). *G status* is the scheduler's view of one G. |
| **wait reason** | A more specific label the runtime substitutes when a G is in `waiting` state (e.g. `chan receive` instead of just `waiting`). |

## How `fmtstack` colors it

| Color           | Highlights                                                |
|-----------------|------------------------------------------------------------|
| bold cyan       | goroutine header                                           |
| bold yellow     | function in **user code**                                  |
| yellow          | function in runtime or dependency                          |
| magenta         | `created by ...` line                                      |
| green           | directory of a user-code file                              |
| gray            | dependency directory, `+0xNN` offsets, standalone lines    |
| bold white      | `file.go:line`                                             |

User code = path that contains **neither** `/usr/local/go/src/` nor `/go/pkg/mod/` (see `emit` in `internal/format/format.go`).
