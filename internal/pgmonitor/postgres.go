/*
 Copyright 2021 - 2023 Crunchy Data Solutions, Inc.
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

package pgmonitor

import (
	"context"
	"strings"

	corev1 "k8s.io/api/core/v1"

	ivory "github.com/ivorysql/ivory-operator/internal/ivory"
	"github.com/ivorysql/ivory-operator/internal/logging"
	"github.com/ivorysql/ivory-operator/pkg/apis/ivory-operator.ivorysql.org/v1beta1"
)

const (
	// MonitoringUser is a Ivory user created by pgMonitor configuration
	MonitoringUser = "ccp_monitoring"
)

// IvorySQLHBAs provides the Ivory HBA rules for allowing the monitoring
// exporter to be accessible
func IvorySQLHBAs(inCluster *v1beta1.IvoryCluster, outHBAs *ivory.HBAs) {
	if ExporterEnabled(inCluster) {
		// Limit the monitoring user to local connections using SCRAM.
		outHBAs.Mandatory = append(outHBAs.Mandatory,
			*ivory.NewHBA().TCP().User(MonitoringUser).Method("scram-sha-256").Network("127.0.0.0/8"),
			*ivory.NewHBA().TCP().User(MonitoringUser).Method("scram-sha-256").Network("::1/128"),
			*ivory.NewHBA().TCP().User(MonitoringUser).Method("reject"))
	}
}

// IvorySQLParameters provides additional required configuration parameters
// that Ivory needs to support monitoring
func IvorySQLParameters(inCluster *v1beta1.IvoryCluster, outParameters *ivory.Parameters) {
	if ExporterEnabled(inCluster) {
		// Exporter expects that shared_preload_libraries are installed
		// pg_stat_statements: https://access.ivorysql.org/documentation/pgmonitor/latest/exporter/
		// pgnodemx: https://github.com/ivorysql/pgnodemx
		libraries := []string{"pg_stat_statements", "pgnodemx"}

		defined, found := outParameters.Mandatory.Get("shared_preload_libraries")
		if found {
			libraries = append(libraries, defined)
		}

		outParameters.Mandatory.Add("shared_preload_libraries", strings.Join(libraries, ","))
		outParameters.Mandatory.Add("pgnodemx.kdapi_path",
			ivory.DownwardAPIVolumeMount().MountPath)
	}
}

// DisableExporterInIvorySQL disables the exporter configuration in IvorySQL.
// Currently the exporter is disabled by removing login permissions for the
// monitoring user.
// TODO: evaluate other uninstall/removal options
func DisableExporterInIvorySQL(ctx context.Context, exec ivory.Executor) error {
	log := logging.FromContext(ctx)

	stdout, stderr, err := exec.Exec(ctx, strings.NewReader(`
		SELECT pg_catalog.format('ALTER ROLE %I NOLOGIN', :'username')
		 WHERE EXISTS (SELECT 1 FROM pg_catalog.pg_roles WHERE rolname = :'username')
		\gexec`),
		map[string]string{
			"username": MonitoringUser,
		})

	log.V(1).Info("monitoring user disabled", "stdout", stdout, "stderr", stderr)

	return err
}

// EnableExporterInIvorySQL runs SQL setup commands in `database` to enable
// the exporter to retrieve metrics. pgMonitor objects are created and expected
// extensions are installed. We also ensure that the monitoring user has the
// current password and can login.
func EnableExporterInIvorySQL(ctx context.Context, exec ivory.Executor,
	monitoringSecret *corev1.Secret, database, setup string) error {
	log := logging.FromContext(ctx)

	stdout, stderr, err := exec.ExecInAllDatabases(ctx,
		strings.Join([]string{
			// Quiet NOTICE messages from IF EXISTS statements.
			// - https://www.ivorysql.org/docs/current/runtime-config-client.html
			`SET client_min_messages = WARNING;`,

			// Exporter expects that extension(s) to be installed in all databases
			// pg_stat_statements: https://access.ivorysql.org/documentation/pgmonitor/latest/exporter/
			"CREATE EXTENSION IF NOT EXISTS pg_stat_statements;",

			// Run idempotent update
			"ALTER EXTENSION pg_stat_statements UPDATE;",
		}, "\n"),
		map[string]string{
			"ON_ERROR_STOP": "on", // Abort when any one statement fails.
			"QUIET":         "on", // Do not print successful commands to stdout.
		},
	)

	log.V(1).Info("applied pgMonitor objects", "database", "current and future databases", "stdout", stdout, "stderr", stderr)

	// NOTE: Setup is run last to ensure that the setup sql is used in the hash
	if err == nil {
		stdout, stderr, err = exec.ExecInDatabasesFromQuery(ctx,
			`SELECT :'database'`,
			strings.Join([]string{
				// Quiet NOTICE messages from IF EXISTS statements.
				// - https://www.ivorysql.org/docs/current/runtime-config-client.html
				`SET client_min_messages = WARNING;`,

				// Setup.sql file from the exporter image. sql is specific
				// to the IvorySQL version
				setup,

				// pgnodemx: https://github.com/ivorysql/pgnodemx
				// The `monitor` schema is hard-coded in the setup SQL files
				// from pgMonitor configuration
				// https://github.com/ivorysql/pgmonitor/blob/master/postgres_exporter/common/queries_nodemx.yml
				"CREATE EXTENSION IF NOT EXISTS pgnodemx WITH SCHEMA monitor;",

				// Run idempotent update
				"ALTER EXTENSION pgnodemx UPDATE;",

				// ccp_monitoring user is created in Setup.sql without a
				// password; update the password and ensure that the ROLE
				// can login to the database
				`ALTER ROLE :"username" LOGIN PASSWORD :'verifier';`,
			}, "\n"),
			map[string]string{
				"database": database,
				"username": MonitoringUser,
				"verifier": string(monitoringSecret.Data["verifier"]),

				"ON_ERROR_STOP": "on", // Abort when any one statement fails.
				"QUIET":         "on", // Do not print successful commands to stdout.
			},
		)

		log.V(1).Info("applied pgMonitor objects", "database", database, "stdout", stdout, "stderr", stderr)
	}

	return err
}
