# User Management

IVYO manages IvorySQL users that you define in IvoryCluster.spec.users.
There, you can list their [role attributes](https://www.postgresql.org/docs/current/role-attributes.html) and which databases they can access.

Below is some information on how the user and database management systems work. To try out some examples, please see the [user and database management](https://github.com/IvorySQL/ivory-operator/blob/master/docs/content/tutorial/user-management.md) section of the tutorial.

## Understanding Default User Management

When you create an Ivory cluster with IVYO and do not specify any additional users or databases, IVYO will do the following:

- Create a database that matches the name of the Ivory cluster.
- Create an unprivileged Ivory user with the name of the cluster. This user has access to the database created in the previous step.
- Create a Secret with the login credentials and connection details for the Ivory user in relation to the database. This is stored in a Secret named `<clusterName>-pguser-<clusterName>`. These credentials include:
  - `user`: The name of the user account.
  - `password`: The password for the user account.
  - `dbname`: The name of the database that the user has access to by default.
  - `host`: The name of the host of the database.
    This references the [Service](https://kubernetes.io/docs/concepts/services-networking/service/) of the primary Ivory instance.
  - `port`: The port that the database is listening on.
  - `uri`: A [IvorySQL connection URI](https://www.postgresql.org/docs/current/libpq-connect.html#LIBPQ-CONNSTRING)
    that provides all the information for logging into the Ivory database.
  - `jdbc-uri`: A [IvorySQL JDBC connection URI](https://jdbc.postgresql.org/documentation/use/)
    that provides all the information for logging into the Ivory database via the JDBC driver.

You can see this default behavior in the [connect to a cluster](https://github.com/IvorySQL/ivory-operator/blob/master/docs/content/tutorial/connect-cluster.md) portion of the tutorial.

As an example, using our `hippo` Ivory cluster, we would see the following created:

- A database named `hippo`.
-An Ivory user named `hippo`.
- A Secret named `hippo-pguser-hippo` that contains the user credentials and connection information.

While the above defaults may work for your application, there are certain cases where you may need to customize your user and databases:

- You may require access to the `ivorysql` superuser.
- You may need to define privileges for your users.
- You may need multiple databases in your cluster, e.g. in a multi-tenant application.
- Certain users may only be able to access certain databases.

## Custom Users and Databases

Users and databases can be customized in the spec.users section of the custom resource. These can be adding during cluster creation and adjusted over time, but it's important to note the following:

- If `spec.users` is set during cluster creation, IVYO will **not** create any default users or databases except for `ivorysql`. If you want additional databases, you will need to specify them.
- For any users added in `spec.users`, IVYO will created a Secret of the format `<clusterName>-pguser-<userName>`. This will contain the user credentials.
  - If no databases are specified, `dbname` and `uri` will not be present in the Secret.
  - If at least one `spec.users.databases` is specified, the first database in the list will be populated into the connection credentials.
- To prevent accidental data loss, IVYO does not automatically drop users. We will see how to drop a user below.
- Similarly, to prevent accidental data loss IVYO does not automatically drop databases. We will see how to drop a database below.
- Role attributes are not automatically dropped if you remove them. You will have to set the inverse attribute to drop them (e.g. `NOSUPERUSER`).
- The special `ivorysql` user can be added as one of the custom users; however, the privileges of the users cannot be adjusted.

For specific examples for how to manage users, please see the [user and database management](https://github.com/IvorySQL/ivory-operator/blob/master/docs/content/tutorial/user-management.md) section of the tutorial.

## Generated Passwords

IVYO generates a random password for each Ivory user it creates. Ivory allows almost any character
in its passwords, but your application may have stricter requirements. To have IVYO generate a password
without special characters, set the `spec.users.password.type` field for that user to `AlphaNumeric`.
For complete control over a user's password, see the [custom passwords](#custom-passwords) section.

To have IVYO generate a new password, remove the existing `password` field from the user _Secret_.
For example, on an Ivory cluster named `hippo` in the `ivory-operator` namespace with
an Ivory user named `hippo`, use the following `kubectl patch` command:

```shell
kubectl patch secret -n ivory-operator hippo-pguser-hippo -p '{"data":{"password":""}}'
```

## Custom Passwords {#custom-passwords}

There are cases where you may want to explicitly provide your own password for an Ivory user.
IVYO determines the password from an attribute in the user Secret called `verifier`. This contains
a hashed copy of your password. When `verifier` changes, IVYO will loa d the contents of the verifier
into your Ivory cluster. This method allows for the secure transmission of the password into the
Ivory database.

Ivory provides two methods for hashing passwords: SCRAM-SHA-256 and MD5.
IVYO uses the preferred (and as of IvorySQL 2, default) method, SCRAM-SHA-256.

There are two ways you can set a custom password for a user. You can provide a plaintext password
in the `password` field and remove the `verifier`. When IVYO detects a password without a verifier
it will generate the SCRAM `verifier` for you. Optionally, you can generate your own password and
verifier. When both values are found in the user secret IVYO will not generate anything. Once the
password and verifier are found IVYO will ensure the provided credential is properly set in postgres.

### Example

For example, let's say we have an Ivory cluster named `hippo` and an Ivory user named `hippo`.
The Secret then would be called `hippo-pguser-hippo`. We want to set the password for `hippo` to
be `datalake` and we can achieve this with a simple `kubectl patch` command. The below assumes that
the Secret is stored in the `ivory-operator` namespace:

```shell
kubectl patch secret -n ivory-operator hippo-pguser-hippo -p \
   '{"stringData":{"password":"datalake","verifier":""}}'
```

{{% notice tip %}}
We can take advantage of the [Kubernetes Secret](https://kubernetes.io/docs/reference/kubernetes-api/config-and-storage-resources/secret-v1/#Secret)
`stringData` field to specify non-binary secret data in string form.
{{% /notice %}}

IVYO generates the SCRAM verifier and applies the updated password to Ivory, and you will be
able to log in with the password `datalake`.
