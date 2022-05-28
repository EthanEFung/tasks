package main

import (
	"encoding/binary"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/boltdb/bolt"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Short: "Tasks is a simple todo application",
	Long:  `A simple cli todo application that will allow you to keep track of things todo`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("welcome to tasks, enter `tasks --help` to see a list of commands")
	},
}

var addCmd = &cobra.Command{
	Use:   "add ...",
	Short: "Add a task",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("starting add transaction")
		task := ""
		for _, arg := range args {
			task += arg + " "
		}
		task = strings.TrimSpace(task)
		db, err := bolt.Open("tasks.db", 0600, &bolt.Options{Timeout: 1 * time.Second})
		if err != nil {
			log.Fatal(err)
		}
		defer db.Close()

		err = db.Update(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte("tasks"))
			id, _ := b.NextSequence()
			err := b.Put(itob(int(id)), []byte(task))
			if err != nil {
				fmt.Errorf("add task err: %s", err)
				return nil
			}
			fmt.Println("Running add task")
			return nil
		})
		if err != nil {
			fmt.Errorf("failed add task exec %s", err)
		}
		fmt.Printf("Added task: %s\n", task)
	},
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "list your tasks",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("starting list transaction")
		db, err := bolt.Open("tasks.db", 0600, &bolt.Options{Timeout: 1 * time.Second})
		if err != nil {
			log.Fatal(err)
		}
		defer db.Close()
		db.View(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte("tasks"))
			var i int
			b.ForEach(func(_, v []byte) error {
				fmt.Printf("%d) %s\n", i+1, v)
				i++
				return nil
			})
			return nil
		})
	},
}

var doCmd = &cobra.Command{
	Use:   "do { task number }",
	Short: "Mark task as done when passed a task number",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// convert the argument to an int to a byte array
		db, err := bolt.Open("tasks.db", 0600, &bolt.Options{Timeout: 1 * time.Second})
		if err != nil {
			log.Fatal(err)
		}
		defer db.Close()

		err = db.Update(func(tx *bolt.Tx) error {
			bucket := tx.Bucket([]byte("tasks"))
			cursor := bucket.Cursor()

			index, err := strconv.Atoi(args[0])

			if err != nil {
				log.Fatal("Please pass an integer")
				return err
			}
			i := 1

			for key, _ := cursor.First(); key != nil; key, _ = cursor.Next() {
				if i == index {
					bucket.Delete(key)
				}
				i++
			}
			return nil
		})
	},
}

func init() {
	db, err := bolt.Open("tasks.db", 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("tasks"))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
	rootCmd.AddCommand(addCmd, listCmd, doCmd)
}

func itob(v int) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(v))
	return b
}
func btoi(k []byte) uint64 {
	return binary.BigEndian.Uint64(k)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(os.Stderr, err)
		os.Exit(1)
	}
}
