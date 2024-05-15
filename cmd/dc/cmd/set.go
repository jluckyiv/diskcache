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
	"time"

	"github.com/jluckyiv/diskcache"
	"github.com/spf13/cobra"
)

// setCmd represents the put command
var setCmd = &cobra.Command{
	Use:   "set",
	Short: "Set a value in the cache",
	Run: func(cmd *cobra.Command, args []string) {
		key, _ := cmd.Flags().GetString("key")
		value, _ := cmd.Flags().GetString("val")
		duration, _ := cmd.Flags().GetDuration("duration")
		cache, err := diskcache.New(cacheDir)
		cobra.CheckErr(err)
		err = cache.Set(key, []byte(value), duration)
		cobra.CheckErr(err)
		fmt.Printf("Set %s=%s for %s\n", key, value, duration)
	},
}

func init() {
	rootCmd.AddCommand(setCmd)
	setCmd.Flags().StringP("key", "k", "", "Key to store the value")
	setCmd.Flags().StringP("val", "v", "", "Value to store")
	setCmd.Flags().DurationP("duration", "d", 1*time.Hour, "Duration to store the value")
	_ = setCmd.MarkFlagRequired("key")
	_ = setCmd.MarkFlagRequired("value")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// putCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// putCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
