package cmd

import (
	"errors"
	"fmt"
	"strings"

	resty "github.com/go-resty/resty/v2"
	helper "github.com/Vioneta/cli/client"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var storeRepositoriesDeleteCmd = &cobra.Command{
	Use:     "delete [slug]",
	Aliases: []string{"del", "remove"},
	Short:   "Delete repository from Home Assistant store",
	Long: `
Remove a repository of add-ons that isn't in use from the Home Assistant store.
`,
	Example: `
ha store delete 94cfad5a
`,
	ValidArgsFunction: storeRepositoriesDeleteCompletions,
	Args:              cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		log.WithField("args", args).Debug("store delete")

		section := "store"
		command := "repositories/{slug}"

		url, err := helper.URLHelper(section, command)
		if err != nil {
			fmt.Println(err)
			ExitWithError = true
			return
		}

		request := helper.GetJSONRequest()

		slug := args[0]
		request.SetPathParams(map[string]string{
			"slug": slug,
		})

		resp, err := request.Delete(url)

		// returns 200 OK or 400, everything else is wrong
		if err == nil {
			if resp.StatusCode() != 200 && resp.StatusCode() != 400 {
				err = errors.New("unexpected server response")
				log.Error(err)
			} else if !resty.IsJSONType(resp.Header().Get("Content-Type")) {
				err = errors.New("API did not return a JSON response")
				log.Error(err)
			}
		}

		if err != nil {
			fmt.Println(err)
			ExitWithError = true
		} else {
			ExitWithError = !helper.ShowJSONResponse(resp)
		}
	},
}

func init() {
	storeCmd.AddCommand(storeRepositoriesDeleteCmd)
}

func storeRepositoriesDeleteCompletions(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) != 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	resp, err := helper.GenericJSONGet("store", "")
	if err != nil || !resp.IsSuccess() {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	var ret []string
	data := resp.Result().(*helper.Response)
	if data.Result == "ok" && data.Data["repositories"] != nil {
		if repos, ok := data.Data["repositories"].([]interface{}); ok {
			for _, repo := range repos {
				var m map[string]interface{}
				if m, ok = repo.(map[string]interface{}); !ok {
					continue
				}
				var s string
				if s, ok = m["slug"].(string); !ok {
					continue
				}
				ret = append(ret, s)
				var ds []string
				if s, ok = m["name"].(string); ok && s != "" {
					ds = append(ds, s)
				}
				if s, ok = m["url"].(string); ok && s != "" {
					ds = append(ds, s)
				}
				if len(ds) != 0 {
					ret[len(ret)-1] += "\t" + strings.Join(ds, ", ")
				}
			}
		}
	}
	return ret, cobra.ShellCompDirectiveNoFileComp
}
