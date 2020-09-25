// Package main contains API for speedtracker CLI
package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	cli "github.com/urfave/cli/v2"
    spinner "github.com/briandowns/spinner"
)

type netData struct {
    pings float64
    avgPing float64
    avgDown float64
    avgUp   float64
    maxPing float64 
    maxDown float64 
    maxUp float64
    timeMaxPing time.Time
    timeMaxDown time.Time
    timeMaxUp   time.Time
    minPing float64
    minDown float64
    minUp float64
    timeMinPing time.Time
    timeMinDown time.Time
    timeMinUp   time.Time
}


func checkSpeedTestExists() bool {
    _, err := exec.LookPath("speedtest")
    if err != nil {
        fmt.Printf("Error: %s\n", err.Error())
        return false
    }
    return true
}

func printData(db netData) {
    cmd := exec.Command("clear")
    cmd.Stdout = os.Stdout
    cmd.Run()
    fmt.Println("\n\nNetwork Data:")
    fmt.Printf("\nAvg Ping: %fms\tAvg Download: %fMbit/s\tAvg Upload: %fMbit/s\n",
    db.avgPing, db.avgDown, db.avgUp)
    fmt.Printf("\nMaximum Stats:\n")
    fmt.Printf("Ping: %fms\t%s\n", db.maxPing, db.timeMaxPing.String())
    fmt.Printf("Download: %fms\t%s\n", db.maxDown, db.timeMaxDown.String())
    fmt.Printf("Upload: %fms\t%s\n", db.maxUp, db.timeMaxUp.String())
    fmt.Printf("\nMinimum Stats:\n")
    fmt.Printf("Ping: %fms\t%s\n", db.minPing, db.timeMinPing.String())
    fmt.Printf("Download: %fms\t%s\n", db.minDown, db.timeMinDown.String())
    fmt.Printf("Upload: %fms\t%s\n", db.minUp, db.timeMinUp.String())
}

func parseData(data string) (ping, down, up float64) {
    lines := strings.Split(data, "\n")
    ping, err := strconv.ParseFloat(strings.Split(lines[0], " ")[1], 64)
    if err != nil {
        panic(err)
    }
    down, err = strconv.ParseFloat(strings.Split(lines[1], " ")[1], 64)
    if err != nil {
        panic(err)
    }
    up, err = strconv.ParseFloat(strings.Split(lines[2], " ")[1], 64)
    if err != nil {
        panic(err)
    }
    return ping, down, up
}

func runTracker(db *netData) error {
    cmd := exec.Command("speedtest", "--simple")
    out, err := cmd.Output()
    if err != nil {
        return err
    }
    now := time.Now()
    ping, down, up := parseData(string(out))
    if ping >= db.maxPing || db.maxPing == -1 {
        db.maxPing = ping
        db.timeMaxPing = now
    }
    if ping <= db.minPing || db.minPing == -1 {
        db.minPing = ping
        db.timeMinPing = now
    }
    if down >= db.maxDown || db.maxDown == -1 {
        db.maxDown = down
        db.timeMaxDown = now
    }
    if down <= db.minDown || db.minDown == -1 {
        db.minDown = down
        db.timeMinDown = now
    }
    if up >= db.maxUp || db.maxUp == -1 {
        db.maxUp = up
        db.timeMaxUp = now
    }
    if up <= db.minUp || db.minUp == -1 {
        db.minUp = up
        db.timeMinUp = now
    }
    pings := db.pings
    db.avgPing = (ping + (db.avgPing*pings)) / (pings+1)
    db.avgDown = (down + (db.avgDown*pings)) / (pings+1)
    db.avgUp = (up + (db.avgUp*pings)) / (pings+1)
    db.pings++

    // print results
    printData(*db)
    return nil
}

func main() {
    // flags
    var qFlag = false
    var secs int

    app := &cli.App{
        Name: "speedtracker",
        Usage: "Analyze your network speed and performance over time.",
        Flags: []cli.Flag{
            &cli.BoolFlag{
                Name: "quiet",
                Aliases: []string{"q"},
                Usage: "Quiet mode silences all output, only maintaining stores.",
                Destination: &qFlag,
            },
            &cli.IntFlag{
                Name: "interval",
                Aliases: []string{"i", "int"},
                Usage: "Specify the interval (seconds) between pings.",
                Destination: &secs,
                Value: 15,
            },
        },
        Action: func(c *cli.Context) error {
            if qFlag {
                fmt.Println("Okay, its quiet time.")
            }
            if !checkSpeedTestExists() {
                fmt.Println("Please install speedtest-cli first by running:")
                fmt.Println("   brew install speedtest-cli")
            }
            // timer
            ticker := time.NewTicker(time.Second * time.Duration(secs))

            // handle SIGTERMS
            osChan := make(chan os.Signal)
            signal.Notify(osChan, os.Interrupt, syscall.SIGTERM)
            
            // define our command
            cmd := exec.Command("speedtest", "--simple")
            cmd.Stdout = os.Stdout
            cmd.Stderr = os.Stderr

            // new data struct
            now := time.Now()
            db := netData{
                0,
                -1, -1, -1, -1, -1, -1,
                now, now, now,
                -1, -1, -1,
                now, now, now,
            }

            s := spinner.New(spinner.CharSets[1], 100*time.Millisecond)
            s.Prefix = "Starting speedtracker... "
            s.Start()
            err := runTracker(&db)
            s.Stop()
            if err != nil {
                return err
            }

            for {
                select {
                    case <-osChan:
                        ticker.Stop()
                        return nil
                    case <-ticker.C:
                        s.Prefix = "Getting New Data... "
                        s.Start()
                        err = runTracker(&db)
                        s.Stop()
                        if err != nil {
                            ticker.Stop()
                            return err
                        }
                }
            }
        },
    }

    err := app.Run(os.Args)
    if err != nil {
        log.Fatal(err)
    }

}
