/*
Copyright © 2024 Jackson Lucky <jack@jacksonlucky.net>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/jluckyiv/diskcache"
	"github.com/spf13/cobra"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List the keys in the cache",
	Run: func(cmd *cobra.Command, args []string) {
		cache, err := diskcache.New(cacheDir)
		cobra.CheckErr(err)
		result, err := cache.List()
		cobra.CheckErr(err)
		if len(result) == 0 {
			fmt.Println("No entries found")
			os.Exit(0)
		}

		// Colors are terminal red, yellow, and green from Tokyo Night theme
		// https://github.com/enkia/tokyo-night-vscode-theme?tab=readme-ov-file#tokyo-night-and-tokyo-night-storm
		// https://github.com/enkia/tokyo-night-vscode-theme?tab=readme-ov-file#tokyo-night-light
		expiredStyle := lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#8c4351", Dark: "#f7768e"})
		almostExpiredStyle := lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#8f5e15", Dark: "#e0af68"})
		currentStyle := lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#33635c", Dark: "#73daca"})
		for _, entry := range result {
			expiryString := entry.Expiry.Local().Format(time.DateTime)
			switch {
			case time.Now().After(entry.Expiry):
				fmt.Printf("%s %s\n", expiredStyle.Render(expiryString), entry.Key)
			case time.Until(entry.Expiry).Minutes() < 5:
				fmt.Printf("%s %s\n", almostExpiredStyle.Render(expiryString), entry.Key)
			default:
				fmt.Printf("%s %s\n", currentStyle.Render(expiryString), entry.Key)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
