/*
 Copyright 2021 - 2023 Highgo Solutions, Inc.
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
	"bytes"
	"context"
	"encoding/json"

	"github.com/ivorysql/ivory-operator/internal/logging"
	"github.com/ivorysql/ivory-operator/pkg/apis/ivory-operator.highgo.com/v1beta1"
)

// WriteUsersInIvorySQL calls exec to create users that do not exist in
// IvorySQL. Once they exist, it updates their options and passwords and
// grants them access to their specified databases. The databases must already
// exist.
func WriteUsersInIvorySQL(
	ctx context.Context, exec Executor,
	users []v1beta1.IvoryUserSpec, verifiers map[string]string,
) error {
	log := logging.FromContext(ctx)

	var err error
	var sql bytes.Buffer

	// Prevent unexpected dereferences by emptying "search_path". The "pg_catalog"
	// schema is still searched, and only temporary objects can be created.
	// - https://www.ivorysql.org/docs/current/runtime-config-client.html#GUC-SEARCH-PATH
	_, _ = sql.WriteString(`SET search_path TO '';`)

	// Fill a temporary table with the JSON of the user specifications.
	// "\copy" reads from subsequent lines until the special line "\.".
	// - https://www.ivorysql.org/docs/current/app-psql.html#APP-PSQL-META-COMMANDS-COPY
	_, _ = sql.WriteString(`
CREATE TEMPORARY TABLE input (id serial, data json);
\copy input (data) from stdin with (format text)
`)
	encoder := json.NewEncoder(&sql)
	encoder.SetEscapeHTML(false)

	for i := range users {
		spec := users[i]

		databases := spec.Databases
		options := spec.Options

		// The "ivory" user must always be a superuser that can login to
		// the "ivory" database.
		if spec.Name == "ivory" {
			databases = append(databases[:0:0], "ivory")
			options = `LOGIN SUPERUSER`
		}

		if err == nil {
			err = encoder.Encode(map[string]interface{}{
				"databases": databases,
				"options":   options,
				"username":  spec.Name,
				"verifier":  verifiers[string(spec.Name)],
			})
		}
	}
	_, _ = sql.WriteString(`\.` + "\n")

	// Create the following objects in a transaction so that permissions are
	// correct before any other session sees them.
	// - https://www.ivorysql.org/docs/current/ddl-priv.html
	_, _ = sql.WriteString(`BEGIN;`)

	// Create users that do not already exist. Permissions are granted later.
	// Roles created this way automatically have the LOGIN option.
	// - https://www.ivorysql.org/docs/current/sql-createuser.html
	_, _ = sql.WriteString(`
SELECT pg_catalog.format('CREATE USER %I',
       pg_catalog.json_extract_path_text(input.data, 'username'))
  FROM input
 WHERE NOT EXISTS (
       SELECT 1 FROM pg_catalog.pg_roles
       WHERE rolname = pg_catalog.json_extract_path_text(input.data, 'username'))
 ORDER BY input.id
\gexec
`)

	// Set any options from the specification. Validation ensures that the value
	// does not contain semicolons.
	// - https://www.ivorysql.org/docs/current/sql-alterrole.html
	_, _ = sql.WriteString(`
SELECT pg_catalog.format('ALTER ROLE %I WITH %s PASSWORD %L',
       pg_catalog.json_extract_path_text(input.data, 'username'),
       pg_catalog.json_extract_path_text(input.data, 'options'),
       pg_catalog.json_extract_path_text(input.data, 'verifier'))
  FROM input ORDER BY input.id
\gexec
`)

	// Grant access to any specified databases.
	// - https://www.ivorysql.org/docs/current/sql-grant.html
	_, _ = sql.WriteString(`
SELECT pg_catalog.format('GRANT ALL PRIVILEGES ON DATABASE %I TO %I',
       pg_catalog.json_array_elements_text(
       pg_catalog.json_extract_path(
       pg_catalog.json_strip_nulls(input.data), 'databases')),
       pg_catalog.json_extract_path_text(input.data, 'username'))
  FROM input ORDER BY input.id
\gexec
`)

	// Commit (finish) the transaction.
	_, _ = sql.WriteString(`COMMIT;`)

	stdout, stderr, err := exec.Exec(ctx, &sql,
		map[string]string{
			"ON_ERROR_STOP": "on", // Abort when any one statement fails.
			"QUIET":         "on", // Do not print successful statements to stdout.
		})

	log.V(1).Info("wrote IvorySQL users", "stdout", stdout, "stderr", stderr)

	return err
}