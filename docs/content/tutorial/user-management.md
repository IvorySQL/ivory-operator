# User / Database Management
IVYO comes with some out-of-the-box conveniences for managing users and databases in your Ivory cluster. However, you may have requirements where you need to create additional users, adjust user privileges or add additional databases to your cluster.

For detailed information for how user and database management works in IVYO, please see the [User Management](https://github.com/IvorySQL/ivory-operator/blob/master/docs/content/architecture/user-management.md) section of the architecture guide.

## Creating a New User

You can create a new user with the following snippet in the `ivorycluster` custom resource. Let's add this to our `hippo` database:

```
spec:
  users:
    - name: rhino
```

You can now apply the changes and see that the new user is created. Note the following:

- The user would only be able to connect to the default `ivory` database.
- The user will not have any connection credentials populated into the `hippo-pguser-rhino` Secret.
- The user is unprivileged.

Let's create a new database named `zoo` that we will let the `rhino` user access:

```
spec:
  users:
    - name: rhino
      databases:
        - zoo
```

Inspect the `hippo-pguser-rhino` Secret. You should now see that the `dbname` and `uri` fields are now populated!

We can set role privileges by using the standard [role attributes](https://www.postgresql.org/docs/current/role-attributes.html) that Ivory provides and adding them to the `spec.users.options`. Let's say we want the rhino to become a superuser (be careful about doling out Ivory superuser privileges!). You can add the following to the spec:

```
spec:
  users:
    - name: rhino
      databases:
        - zoo
      options: "SUPERUSER"
```

There you have it: we have created a Ivory user named `rhino` with superuser privileges that has access to the `rhino` database (though a superuser has access to all databases!).

## Adjusting Privileges

Let's say you want to revoke the superuser privilege from `rhino`. You can do so with the following:

```
spec:
  users:
    - name: rhino
      databases:
        - zoo
      options: "NOSUPERUSER"
```

If you want to add multiple privileges, you can add each privilege with a space between them in `options`, e.g.:

```
spec:
  users:
    - name: rhino
      databases:
        - zoo
      options: "CREATEDB CREATEROLE"
```

## Managing the `ivory` User

By default, IVYO does not give you access to the `ivory` user. However, you can get access to this account by doing the following:

```
spec:
  users:
    - name: ivory
```

This will create a Secret of the pattern `<clusterName>-pguser-ivory` that contains the credentials of the `ivory` account. For our `hippo` cluster, this would be `hippo-pguser-ivory`.

## Deleting a User

IVYO does not delete users automatically: after you remove the user from the spec, it will still exist in your cluster. To remove a user and all of its objects, as a superuser you will need to run [`DROP OWNED`](https://www.postgresql.org/docs/current/sql-drop-owned.html) in each database the user has objects in, and [`DROP ROLE`](https://www.postgresql.org/docs/current/sql-droprole.html)
in your Ivory cluster.

For example, with the above `rhino` user, you would run the following:

```
DROP OWNED BY rhino;
DROP ROLE rhino;
```

Note that you may need to run `DROP OWNED BY rhino CASCADE;` based upon your object ownership structure -- be very careful with this command!

## Deleting a Database

IVYO does not delete databases automatically: after you remove all instances of the database from the spec, it will still exist in your cluster. To completely remove the database, you must run the [`DROP DATABASE`](https://www.postgresql.org/docs/current/sql-dropdatabase.html)
command as a Ivory superuser.

For example, to remove the `zoo` database, you would execute the following:

```
DROP DATABASE zoo;
```

## Next Steps

Let's look at how IVYO handles [disaster recovery](https://github.com/IvorySQL/ivory-operator/blob/master/docs/content/tutorial/disaster-recovery.md)!