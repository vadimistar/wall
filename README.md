# Wall

A funny programming langauge, translated to C

## Hello World

main.wall:

```
extern fun puts(msg *char) int

fun main() int32 {
    puts("Hello, World!") 
    return 0
}
```

```
go run cmd/wallc/main.go mall.wall > main.c
gcc main.c -o ./main
./main
```

## Inline C

```
fun main() int32 {
    var c = 50
    inlineC("c *= c")
    return c
}
```

