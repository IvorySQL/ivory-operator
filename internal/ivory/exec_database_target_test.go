/*
 Copyright 2026 IvorySQL Contributors.
 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

 http://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

package ivory

import (
	"context"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
)

// Exercise the actual production Bash loop. The psql command boundary supplies
// catalog names and records target invocations; it does not emulate a database
// or claim that SQL, extensions, authentication, or libpq were exercised.
func TestExecutorDatabaseTargetCommandBoundary(t *testing.T) {
	bash := os.Getenv("IVORY_TEST_BASH")
	if bash == "" {
		var err error
		bash, err = exec.LookPath("bash")
		if err != nil {
			t.Fatal("this command-boundary test requires Bash: ", err)
		}
	}

	// A catalog query must retain its explicit maintenance database. Target
	// queries must use PGDATABASE without an overriding database argument.
	const boundary = `
psql() {
	local arg catalog=false
	for arg in "$@"; do
		if [[ "$arg" == -Aqt ]]; then catalog=true; fi
	done
	local sql
	sql=$(< /dev/stdin)
	if "$catalog"; then
		[[ "${1-}" == -d && "${2-}" == postgres ]] || return 65
		[[ "$sql" == *pg_catalog.pg_database* ]] || return 66
		printf '%s\n' postgres template1 'application db' 'host=127.0.0.1'
		return
	fi
	[[ "$sql" == 'SELECT current_database();' ]] || return 67
	for arg in "$@"; do
		case "$arg" in
			-d|--dbname|--dbname=*)
				printf 'target query overrides PGDATABASE with argument %s\n' "$arg" >&2
				return 64 ;;
		esac
	done
	printf '%s\n' "$PGDATABASE"
}
`
	file := filepath.Join(t.TempDir(), "psql-boundary.bash")
	if err := os.WriteFile(file, []byte(boundary), 0o600); err != nil {
		t.Fatal(err)
	}

	for _, tc := range []struct {
		name string
		vars map[string]string
	}{
		{name: "PostGIS and pgAudit directory query", vars: map[string]string{"ON_ERROR_STOP": "on", "QUIET": "on"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fn := Executor(func(ctx context.Context, stdin io.Reader, stdout, stderr io.Writer, command ...string) error {
				if len(command) < 4 || command[0] != "bash" {
					t.Fatalf("expected production Bash invocation, got %q", command)
				}
				// Use the generated production script unchanged. BASH_ENV installs
				// only the command-boundary fixture before Bash executes it.
				cmd := exec.CommandContext(ctx, bash, command[1:]...)
				env := os.Environ()
				for i := len(env) - 1; i >= 0; i-- {
					if strings.HasPrefix(strings.ToUpper(env[i]), "BASH_ENV=") {
						env = append(env[:i], env[i+1:]...)
					}
				}
				cmd.Env = append(env, "BASH_ENV="+filepath.ToSlash(file))
				// A regular input file also gives Git Bash on Windows a usable
				// /dev/stdin handle; native Go pipes do not provide one there.
				input, err := os.CreateTemp(t.TempDir(), "sql-input-")
				if err != nil {
					return err
				}
				defer input.Close()
				if _, err := io.Copy(input, stdin); err != nil {
					return err
				}
				if _, err := input.Seek(0, io.SeekStart); err != nil {
					return err
				}
				cmd.Stdin, cmd.Stdout, cmd.Stderr = input, stdout, stderr
				return cmd.Run()
			})
			out, errout, err := fn.ExecInAllDatabases(context.Background(), "SELECT current_database();", tc.vars)
			if err != nil {
				t.Fatalf("database target command contract failed: %v; stderr=%s", err, errout)
			}
			want := "postgres\ntemplate1\napplication db\nhost=127.0.0.1\n"
			if out != want {
				t.Fatalf("target invocations = %q, want %q", out, want)
			}
		})
	}
}

// Capture the production Executor boundary without inventing psql connection
// output. In particular, pgMonitor supplies a connect meta-command rather than
// a SELECT of names; the existing special script must remain byte-for-byte.
func TestExecutorLegacyMonitoringQueryCompatibility(t *testing.T) {
	const legacyMonitoringScript = `
sql_target=$(< /dev/stdin)
sql_databases="$1"
shift 1
		
databases=$(psql -d postgres "$@" -Xw -Aqt --file=- <<< "${sql_databases}")
while IFS= read -r database; do
	PGDATABASE="${database}" psql -d postgres "$@" -Xw --file=- <<< "${sql_target}"
done <<< "${databases}"
`
	const catalogQuery = "SET search_path = '';SELECT datname FROM pg_catalog.pg_database" +
		" WHERE datallowconn AND datname NOT IN ('template0')"
	for _, tc := range []struct {
		name         string
		query        string
		allDatabases bool
		vars         map[string]string
	}{
		{
			name:  "pgMonitor connect meta-command",
			query: `\c postgres`,
			vars: map[string]string{
				"database": "postgres", "username": "ccp_monitoring", "verifier": "test-verifier",
				"ON_ERROR_STOP": "on", "QUIET": "on",
			},
		},
		{
			name:         "PgBouncer catalog caller",
			query:        catalogQuery,
			allDatabases: true,
			vars: map[string]string{
				"namespace": "pgbouncer", "username": "_highgopgbouncer", "verifier": "test-verifier",
				"ON_ERROR_STOP": "on", "QUIET": "on",
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			sentinel := errors.New("command capture only")
			calls := 0
			const sql = "SELECT current_database();"
			fn := Executor(func(_ context.Context, stdin io.Reader, stdout, stderr io.Writer, command ...string) error {
				calls++
				body, err := io.ReadAll(stdin)
				if err != nil || string(body) != sql {
					t.Fatalf("SQL input = %q, error = %v", body, err)
				}
				want := []string{"bash", "-ceu", "--", legacyMonitoringScript, "-", tc.query}
				var assignments []string
				for key, value := range tc.vars {
					assignments = append(assignments, "--set="+key+"="+value)
				}
				sort.Strings(assignments)
				want = append(want, assignments...)
				if !reflect.DeepEqual(command, want) {
					t.Fatalf("legacy command changed:\n got %q\nwant %q", command, want)
				}
				_, _ = io.WriteString(stdout, "captured stdout")
				_, _ = io.WriteString(stderr, "captured stderr")
				return sentinel
			})
			var out, errout string
			var err error
			if tc.allDatabases {
				out, errout, err = fn.ExecInAllDatabases(context.Background(), sql, tc.vars)
			} else {
				out, errout, err = fn.ExecInDatabasesFromQuery(context.Background(), tc.query, sql, tc.vars)
			}
			if calls != 1 || err != sentinel || out != "captured stdout" || errout != "captured stderr" {
				t.Fatalf("boundary result: calls=%d stdout=%q stderr=%q err=%v", calls, out, errout, err)
			}
		})
	}
}
