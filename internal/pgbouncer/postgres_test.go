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

package pgbouncer

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/onsi/gomega"
	"gotest.tools/v3/assert"
	corev1 "k8s.io/api/core/v1"
)

func TestSQLAuthenticationQuery(t *testing.T) {
	assert.Equal(t, sqlAuthenticationQuery("some.fn_name"),
		`CREATE OR REPLACE FUNCTION some.fn_name(username TEXT)
RETURNS TABLE(username TEXT, password TEXT) AS '
  SELECT rolname::TEXT, rolpassword::TEXT
  FROM pg_catalog.pg_authid
  WHERE pg_authid.rolname = $1
    AND pg_authid.rolcanlogin
    AND NOT pg_authid.rolsuper
    AND NOT pg_authid.rolreplication
    AND pg_authid.rolname <> ''_highgopgbouncer''
    AND (pg_authid.rolvaliduntil IS NULL OR pg_authid.rolvaliduntil >= CURRENT_TIMESTAMP)'
LANGUAGE SQL STABLE SECURITY DEFINER;`)
}

func TestDisableInIvorySQL(t *testing.T) {
	expected := errors.New("whoops")

	// The first call is to drop objects.
	t.Run("call1", func(t *testing.T) {
		call1 := func(
			_ context.Context, stdin io.Reader, stdout, stderr io.Writer, command ...string,
		) error {
			assert.Assert(t, stdout != nil, "should capture stdout")
			assert.Assert(t, stderr != nil, "should capture stderr")
			assert.Assert(t, strings.Contains(strings.Join(command, "\n"),
				`SELECT datname FROM pg_catalog.pg_database`,
			), "expected all databases and templates")

			b, err := io.ReadAll(stdin)
			assert.NilError(t, err)
			assert.Equal(t, string(b), strings.TrimSpace(`
SET client_min_messages = WARNING;
BEGIN;
DROP FUNCTION IF EXISTS :"namespace".get_auth(username TEXT);
DROP SCHEMA IF EXISTS :"namespace" CASCADE;
SELECT pg_catalog.format('DROP OWNED BY %I CASCADE', :'username')
 WHERE EXISTS (SELECT 1 FROM pg_catalog.pg_roles WHERE rolname = :'username')
\gexec
COMMIT;`))
			gomega.NewWithT(t).Expect(command).To(gomega.ContainElements(
				`--set=namespace=pgbouncer`,
				`--set=username=_highgopgbouncer`,
			), "expected query parameters")

			return expected
		}

		calls := 0
		exec := func(
			ctx context.Context, stdin io.Reader, stdout, stderr io.Writer, command ...string,
		) error {
			calls++
			return call1(ctx, stdin, stdout, stderr, command...)
		}

		ctx := context.Background()
		assert.Equal(t, expected, DisableInIvorySQL(ctx, exec))
		assert.Equal(t, calls, 1, "expected an exec error to return early")
	})

	// The second call is to drop the user.
	t.Run("call2", func(t *testing.T) {
		call2 := func(
			_ context.Context, stdin io.Reader, stdout, stderr io.Writer, command ...string,
		) error {
			assert.Assert(t, stdout != nil, "should capture stdout")
			assert.Assert(t, stderr != nil, "should capture stderr")
			gomega.NewWithT(t).Expect(command).To(gomega.ContainElement(
				`SELECT pg_catalog.current_database()`,
			), "expected the default database")

			b, err := io.ReadAll(stdin)
			assert.NilError(t, err)
			assert.Equal(t, string(b), `SET client_min_messages = WARNING; DROP ROLE IF EXISTS :"username";`)
			gomega.NewWithT(t).Expect(command).To(gomega.ContainElements(
				`--set=username=_highgopgbouncer`,
			), "expected query parameters")

			return expected
		}

		calls := 0
		exec := func(
			ctx context.Context, stdin io.Reader, stdout, stderr io.Writer, command ...string,
		) error {
			calls++
			if calls == 1 {
				return nil
			}
			return call2(ctx, stdin, stdout, stderr, command...)
		}

		ctx := context.Background()
		assert.Equal(t, expected, DisableInIvorySQL(ctx, exec))
		assert.Equal(t, calls, 2, "expected two calls to exec")
	})
}

func TestEnableInIvorySQL(t *testing.T) {
	expected := errors.New("whoops")
	secret := new(corev1.Secret)
	secret.Data = map[string][]byte{
		"pgbouncer-verifier": []byte("digest$and==:whatnot"),
	}

	exec := func(
		_ context.Context, stdin io.Reader, stdout, stderr io.Writer, command ...string,
	) error {
		assert.Assert(t, stdout != nil, "should capture stdout")
		assert.Assert(t, stderr != nil, "should capture stderr")
		assert.Assert(t, strings.Contains(strings.Join(command, "\n"),
			`SELECT datname FROM pg_catalog.pg_database`,
		), "expected all databases and templates")

		b, err := io.ReadAll(stdin)
		assert.NilError(t, err)
		assert.Equal(t, string(b), strings.TrimSpace(`
SET client_min_messages = WARNING;
BEGIN;
SELECT pg_catalog.format('CREATE ROLE %I NOLOGIN', :'username')
 WHERE NOT EXISTS (SELECT 1 FROM pg_catalog.pg_roles WHERE rolname = :'username')
\gexec
SELECT pg_catalog.format('REVOKE ALL PRIVILEGES ON SCHEMA %I FROM %I', nspname, :'username')
  FROM pg_catalog.pg_namespace
 WHERE pg_catalog.has_schema_privilege(:'username', oid, 'CREATE, USAGE')
   AND nspname NOT IN ('pg_catalog', :'namespace')
\gexec
CREATE SCHEMA IF NOT EXISTS :"namespace";
REVOKE ALL PRIVILEGES
    ON SCHEMA :"namespace" FROM PUBLIC, :"username";
 GRANT USAGE
    ON SCHEMA :"namespace" TO :"username";
CREATE OR REPLACE FUNCTION :"namespace".get_auth(username TEXT)
RETURNS TABLE(username TEXT, password TEXT) AS '
  SELECT rolname::TEXT, rolpassword::TEXT
  FROM pg_catalog.pg_authid
  WHERE pg_authid.rolname = $1
    AND pg_authid.rolcanlogin
    AND NOT pg_authid.rolsuper
    AND NOT pg_authid.rolreplication
    AND pg_authid.rolname <> ''_highgopgbouncer''
    AND (pg_authid.rolvaliduntil IS NULL OR pg_authid.rolvaliduntil >= CURRENT_TIMESTAMP)'
LANGUAGE SQL STABLE SECURITY DEFINER;
REVOKE ALL PRIVILEGES
    ON FUNCTION :"namespace".get_auth(username TEXT) FROM PUBLIC, :"username";
 GRANT EXECUTE
    ON FUNCTION :"namespace".get_auth(username TEXT) TO :"username";
ALTER ROLE :"username" SET search_path TO :'namespace';
ALTER ROLE :"username" LOGIN PASSWORD :'verifier';
COMMIT;`))

		gomega.NewWithT(t).Expect(command).To(gomega.ContainElements(
			`--set=namespace=pgbouncer`,
			`--set=username=_highgopgbouncer`,
			`--set=verifier=digest$and==:whatnot`,
		), "expected query parameters")

		return expected
	}

	ctx := context.Background()
	assert.Equal(t, expected, EnableInIvorySQL(ctx, exec, secret))
}

func TestIvorySQLHBAs(t *testing.T) {
	rules := ivorysqlHBAs()
	assert.Equal(t, len(rules), 2)
	assert.Equal(t, rules[0].String(), `hostssl all "_highgopgbouncer" all scram-sha-256`)
	assert.Equal(t, rules[1].String(), `host all "_highgopgbouncer" all reject`)
}
