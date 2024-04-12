package cmd

import (
	"os"
	"text/template"

	"gitbub.com/wbuntu/gin-template/internal/pkg/config"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

const configTemplate = `[general]
# Log level
# trace=6, debug=5, info=4, warning=3, error=2, fatal=1, panic=0
log_level={{ .General.LogLevel }}
log_format="{{ .General.LogFormat }}"
enable_db={{ .General.EnableDB }}
enable_kv_db={{ .General.EnableKVDB }}
enable_leader_election={{ .General.EnableLeaderElection }}

[api]
# ip:port to bind the api server
addr="{{ .API.Addr }}"
# ip:port to bind the api server which enable tls
tls_addr="{{ .API.TLSAddr }}"
# tls crt and key
tls_crt="{{ .API.TLSCrt }}"
tls_key="{{ .API.TLSKey }}"

[db]
# db connection info
type="{{ .DB.Type }}"
dsn="{{ .DB.DSN }}"
min_idle_conns={{ .DB.MinIdleConns }}
max_active_conns={{ .DB.MaxActiveConns }}
conn_lifetime="{{ .DB.ConnLifetime }}"
conn_idletime="{{ .DB.ConnIdletime }}"

[kv_db]
# kv_db connection info
type="{{ .KVDB.Type }}"
dsn="{{ .KVDB.DSN }}"
min_idle_conns={{ .KVDB.MinIdleConns }}
max_active_conns={{ .KVDB.MaxActiveConns }}
conn_lifetime="{{ .KVDB.ConnLifetime }}"
conn_idletime="{{ .KVDB.ConnIdletime }}"
`

var configCmd = &cobra.Command{
	Use:   "configfile",
	Short: "Print ths gin-template configuration file",
	RunE: func(cmd *cobra.Command, args []string) error {
		t := template.Must(template.New("config").Parse(configTemplate))
		err := t.Execute(os.Stdout, &config.C)
		if err != nil {
			return errors.Wrap(err, "execute config template")
		}
		return nil
	},
}
