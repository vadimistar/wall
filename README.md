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
    c := 50
    inlineC("c *= c")
    return c
}
```

# Methods

```
struct Rect {
    width int32 
    height int32
}

fun Rect.area() int32 {
    return .width * .height
}

fun Rect.perim() int32 {
    return 2 * .width + 2 * .height
}

extern fun printf(fmt *char, n int32) int

fun main() int32 {
    r := Rect { width: 10, height: 5 }
    printf("area: %d\n", r.area())
    printf("perim: %d\n", r.perim())
    return 0
}
```
