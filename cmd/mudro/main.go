package main

import (
	"fmt"
	"os"

	claudeusageproxyApp "github.com/goritskimihail/mudro/tools/maintenance/claudeusageproxy/app"
	commentbackfillApp "github.com/goritskimihail/mudro/tools/backfill/commentbackfill/app"
	importerApp "github.com/goritskimihail/mudro/tools/importers/importer/app"
	mediabackfillApp "github.com/goritskimihail/mudro/tools/backfill/mediabackfill/app"
	mementoApp "github.com/goritskimihail/mudro/tools/maintenance/memento/app"
	s3backfillApp "github.com/goritskimihail/mudro/tools/backfill/s3backfill/app"
	tgcommentmediaimportApp "github.com/goritskimihail/mudro/tools/importers/tgcommentmediaimport/app"
	tgcommentscsvimportApp "github.com/goritskimihail/mudro/tools/importers/tgcommentscsvimport/app"
	tgcommentsimportApp "github.com/goritskimihail/mudro/tools/importers/tgcommentsimport/app"
	tgcsvimportApp "github.com/goritskimihail/mudro/tools/importers/tgcsvimport/app"
	tgdedupeApp "github.com/goritskimihail/mudro/tools/maintenance/tgdedupe/app"
	tghtmlimportApp "github.com/goritskimihail/mudro/tools/importers/tghtmlimport/app"
	tgimportApp "github.com/goritskimihail/mudro/tools/importers/tgimport/app"
	tgloadApp "github.com/goritskimihail/mudro/tools/importers/tgload/app"
	tgrootmergeApp "github.com/goritskimihail/mudro/tools/maintenance/tgrootmerge/app"
	vkimportApp "github.com/goritskimihail/mudro/tools/importers/vkimport/app"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Mudro CLI Tool")
		fmt.Println("Usage: mudro <command> [args...]")
		fmt.Println("\nAvailable commands:")
		fmt.Println("  claudeusageproxy       - Claude usage proxy")
		fmt.Println("  commentbackfill        - Backfill comments")
		fmt.Println("  importer               - Generic importer")
		fmt.Println("  mediabackfill          - Backfill media")
		fmt.Println("  memento                - Memento maintenance")
		fmt.Println("  s3backfill             - Backfill S3 buckets")
		fmt.Println("  tgcommentmediaimport   - Import TG comment media")
		fmt.Println("  tgcommentscsvimport    - Import TG comments from CSV")
		fmt.Println("  tgcommentsimport       - Import TG comments")
		fmt.Println("  tgcsvimport            - Import TG CSV data")
		fmt.Println("  tgdedupe               - Deduplicate TG data")
		fmt.Println("  tghtmlimport           - Import TG HTML")
		fmt.Println("  tgimport               - TG standard import")
		fmt.Println("  tgload                 - TG loader tool")
		fmt.Println("  tgrootmerge            - TG root merger")
		fmt.Println("  vkimport               - VK standard import")
		os.Exit(1)
	}

	command := os.Args[1]

	// Shift arguments so subcommands parsing flag.Args() receive their flags natively.
	os.Args = append([]string{os.Args[0] + " " + command}, os.Args[2:]...)

	switch command {
	case "claudeusageproxy":
		claudeusageproxyApp.Run()
	case "commentbackfill":
		commentbackfillApp.Run()
	case "importer":
		importerApp.Run()
	case "mediabackfill":
		mediabackfillApp.Run()
	case "memento":
		mementoApp.Run()
	case "s3backfill":
		s3backfillApp.Run()
	case "tgcommentmediaimport":
		tgcommentmediaimportApp.Run()
	case "tgcommentscsvimport":
		tgcommentscsvimportApp.Run()
	case "tgcommentsimport":
		tgcommentsimportApp.Run()
	case "tgcsvimport":
		tgcsvimportApp.Run()
	case "tgdedupe":
		tgdedupeApp.Run()
	case "tghtmlimport":
		tghtmlimportApp.Run()
	case "tgimport":
		tgimportApp.Run()
	case "tgload":
		tgloadApp.Run()
	case "tgrootmerge":
		tgrootmergeApp.Run()
	case "vkimport":
		vkimportApp.Run()
	default:
		fmt.Printf("Unknown command: %s\n", command)
		os.Exit(1)
	}
}
