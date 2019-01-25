package node

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/blocklayerhq/chainkit/ui"
	"github.com/blocklayerhq/chainkit/util"
	"github.com/manifoldco/promptui"
	"github.com/sergi/go-diff/diffmatchpatch"
)

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	if _, err = io.Copy(out, in); err != nil {
		return err
	}
	out.Sync()
	return nil
}

func spawnGenesisEditor(ctx context.Context, genesisPath string) error {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "nano"
	}

	newPath := genesisPath + ".edit"
	if err := copyFile(genesisPath, newPath); err != nil {
		return err
	}

	if err := util.Run(ctx, editor, newPath); err != nil {
		return err
	}

	dataOld, err := ioutil.ReadFile(genesisPath)
	if err != nil {
		return err
	}
	dataNew, err := ioutil.ReadFile(newPath)
	if err != nil {
		return err
	}

	dmp := diffmatchpatch.New()
	diffs := dmp.DiffMain(string(dataOld), string(dataNew), false)
	if len(diffs) <= 0 {
		ui.Info("No changes detected, ignoring the edits")
	}
	fmt.Println(dmp.DiffPrettyText(diffs))

	msgs := []string{"Yes, apply the changes", "No, keep the original", "Abort the start"}
	prompt := promptui.Select{
		Label: "Do you want to start the chain with this genesis file?",
		Items: msgs,
	}
	_, result, err := prompt.Run()
	if err != nil {
		return err
	}

	switch result {
	case msgs[0]:
		ui.Info("Applying the diff")
		if err := os.Rename(newPath, genesisPath); err != nil {
			return err
		}
	case msgs[1]:
		ui.Info("Starting the chain with the original genesis file (ignoring the changes)")
	case msgs[2]:
		ui.Fatal("Aborting the start per user request (chain is already initialized, if you need to reset: rm -rf ./state)")
	}
	return nil
}
