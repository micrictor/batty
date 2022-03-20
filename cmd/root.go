package cmd

import (
	"bufio"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strings"

	"github.com/micrictor/batty/internal/tty"
	"github.com/micrictor/typo-generator/pkg/mapping"
	"github.com/spf13/cobra"
)

var (
	keyMapping *mapping.Mapping
	rate       float32
	rootCmd    = &cobra.Command{
		Use:   "batty /dev/ttyX",
		Short: "Make ttys drive people batty",
		Long:  `Randomly introduce typos to an open tty on the save device`,
		Run:   cmdRun,
	}
)

const backspace = byte(8)

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().Float32P("rate", "r", 0.1, "Typo induction rate for alphabetical characters")
	rootCmd.Flags().StringP("layout", "l", "qwerty", "Keyboard layout to generate typos for")
}

func cmdRun(cmd *cobra.Command, args []string) {
	var err error
	layout, _ := cmd.Flags().GetString("layout")
	rate, _ = cmd.Flags().GetFloat32("rate")
	keyMapping, err = mapping.New(layout)
	if err != nil {
		log.Fatalf("Failed to get keyboard map for layout!\n%v", err)
	}

	if len(args) < 1 {
		log.Fatal("Expected path to TTY as argument")
	}
	ttyPath := args[0]

	t, err := tty.New(ttyPath)
	if err != nil {
		log.Fatalf("Failed to get tty for read/write: %v", err)
	}

	t.Hook(typoHook)

	fmt.Print("TTY hooked, press q to exit")
	bufio.NewReader(os.Stdin).ReadBytes('q')
}

func typoHook(inputCharacter rune) []byte {
	var isCapital bool
	normalizedCharacter := rune(strings.ToLower(string(inputCharacter))[0])
	if normalizedCharacter != inputCharacter {
		isCapital = true
	} else {
		isCapital = false
	}

	// Only operate on alphabetical characters
	if normalizedCharacter < 97 || normalizedCharacter > 122 {
		return []byte{}
	}

	// Randomly decide if we should operate on this character
	randomFloat := rand.Float32()
	if randomFloat > rate {
		return []byte{}
	}

	possibleTypos, err := keyMapping.FindTypos(normalizedCharacter)
	if err != nil {
		log.Printf("Failed to get typo: %v", err)
		return []byte{}
	}

	selectedTypo := possibleTypos[rand.Intn(len(possibleTypos)-1)]
	log.Printf("Selected typo %c for input %c", selectedTypo, inputCharacter)

	if isCapital {
		outputCharacter := strings.ToUpper(string(selectedTypo))[0]
		return []byte{backspace, outputCharacter}
	}

	return []byte{backspace, byte(selectedTypo)}
}
