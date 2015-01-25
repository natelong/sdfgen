package main

import (
    "fmt"
    "log"
    "flag"
    "os"
    "path"
    "strings"
    "time"

    "image"
    "image/color"
    "image/gif"
    "image/jpeg"
    "image/png"
)

var read   []int
var write  []float64
var w      int
var h      int
var min    int
var max    int
var spread int = 20
var outSize int = 128
var scale int = 0

func main() {
    flag.Parse()

    inFileName   := flag.Arg(0)
    outFileName  := flag.Arg(1)

    if inFileName == "" {
        log.Fatal("Need input file name to process")
    }

    if outFileName == "" {
        log.Fatal("Need output file name to write to")
    }


    ext := strings.ToLower(path.Ext(outFileName))

    reader, err := os.Open(inFileName)
    if err != nil {
        log.Fatal(err)
    }
    defer reader.Close()



    start := time.Now()

    img, _, err := image.Decode(reader)
    if err != nil {
        log.Fatal(err)
    }

    bounds := img.Bounds()
    w = bounds.Max.X
    h = bounds.Max.Y

    if w % outSize != 0 || h % outSize != 0 {
        log.Fatalf("Image size must be evenly divisible by %v", outSize)
    }
    scale = int(w / outSize)

    read  = make([]int, w*h)
    write = make([]float64, w*h)

    for y := 0; y < h; y++ {
        for x := 0; x < w; x++ {
            _, _, _, a := img.At(x, y).RGBA()

            if a < uint32(65534/2) {
                read[w*y + x] = 0
            } else {
                read[w*y + x] = 1
            }
        }
    }

    max = 0
    min = w*h

    for y := 0; y < h; y++ {
        for x := 0; x < w; x++ {
            n := nearest(x, y)
            if n > max { max = n }
            if n < min { min = n }
            write[w*y+x] = float64(n)
        }
    }

    // Normalize
    for i := range write {
        write[i] = (write[i] - float64(min)) / float64(max - min)
    }

    finalImage := image.NewGray(image.Rect(0, 0, outSize, outSize))
    var c color.Gray

    for y := 0; y < outSize; y++ {
        for x := 0; x < outSize; x++ {
            a := avg(x, y)
            c.Y = uint8(a * 255)
            finalImage.Set(x, y, c)
        }
    }



    writer, err := os.Create(outFileName)
    if err != nil {
        log.Fatal(err)
    }
    defer writer.Close()

    switch ext {
        case ".png":
            png.Encode(writer, finalImage)
        case ".gif":
            gif.Encode(writer, finalImage, &gif.Options{256, nil, nil})
        case ".jpg":
            fallthrough
        case ".jpeg":
            jpeg.Encode(writer, finalImage, &jpeg.Options{100})
        default:
            log.Fatalf("Couldn't output image type: %v", ext)
    }

    end := time.Now()
    fmt.Printf("Processing time: %v\n", end.Sub(start))
}

func nearest(x, y int) int {
    var dx int
    var dy int
    var t  int

    if at(x, y) == 1 {
        t = 0
    } else {
        t = 1
    }

    min := spread

    Outer:
    for i := 1; i < spread; i++ {
        dy = y - i
        for dx = x-i; dx <= x+i; dx++ {
            if at(dx, dy) == t && i < min {
                min = i
                break Outer
            }
        }

        dy = y + i
        for dx = x-i; dx <= x+i; dx++ {
            if at(dx, dy) == t && i < min {
                min = i
                break Outer
            }
        }

        dx = x - i
        for dy = y-i+1; dy <= y+i-1; dy++ {
            if at(dx, dy) == t && i < min {
                min = i
                break Outer
            }
        }

        dx = x + i
        for dy = y-i+1; dy <= y+i-1; dy++ {
            if at(dx, dy) == t && i < min {
                min = i
                break Outer
            }
        }
    }

    if at(x, y) == 1 {
        return min
    } else {
        return -min
    }
}

func at(x, y int) int {
    if x >= 0 && x < w && y >= 0 && y < h {
        return read[w*y+x]
    } else {
        return -1
    }
}

func atw(x, y int) float64 {
    if x >= 0 && x < w && y >= 0 && y < h {
        return write[w*y+x]
    } else {
        return 0
    }
}

func avg(x, y int) float64 {
    var count int
    var total float64

    for dy := 0; dy < scale; dy++ {
        for dx := 0; dx < scale; dx++ {
            total += atw(x*scale+dx, y*scale+dy)
            count++
        }
    }

    return total / float64(count)
}

func normalize(val float64) float64 {
    return (val - float64(min)) / float64(max - min)
}