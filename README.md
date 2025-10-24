# Demo app to show how DB Branching in mirrord works

1. Create a cluster `k3d cluster create db-branching-mirrord`
1. Install the mirrord operator on the cluster: `helm install -f values.yaml mirrord-operator metalbear/mirrord-operator` we have already enabled the db branching feature in the values.yaml
1. Apply the k8s.yaml file `kubectl apply -f k8s.yaml`
1. Run using mirrord
1. Send a request to add a user: `curl -X POST "http://localhost:8080/users?name=Alice"`
1. Confirm that the user was added in the DB:
    ```
    $ MYSQL_POD=$(kubectl get pod -l app=mysql -o jsonpath='{.items[0].metadata.name}')
    $ kubectl exec -it $MYSQL_POD -- mysql -uroot -prootpassword -e "USE users_db; SELECT * FROM users;"
    ```
1. Add this to mirrord.json:
    ```json
    {
        "db_branches": [
            {
            "id": "branch-test-01",
            "type": "mysql",
            "version": "8.0",
            "name": "users_db",
            "ttl_secs": 300,
            "connection": {
                "url": {
                "type": "env",
                "variable": "DB_CONNECTION_URL"
                }
            }
            }
        ]
    }
    ```

1. Change code:
    ```go
    _, err = db.Exec(`CREATE TABLE IF NOT EXISTS users (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255),
    email VARCHAR(255)
    )`)

    other code to show email also
    ```

1. Run again using mirrord

1. send a request with a new name and email:
    ```
    $ curl -X POST "http://localhost:8080/users?name=Bob&email=bob@example.com"
    ok: inserted "Bob" into users_db.users
    ```

1. Confirm user was added:
    ```
    $ curl "http://localhost:8080/users" 
    1       Bob bob@example.com
    ```
1. Show that mirrord created a branch of the original DB where this was stored instead of the original DB:
    ```
    ❯ k get pods
NAME                                READY   STATUS    RESTARTS   AGE
mirrord-mysql-branch-db-pod-zq4rs   1/1     Running   0          3m45s
mysql-854884d79b-mgmvj              1/1     Running   0          109m
users-api-5c8579d7dd-98lgz          1/1     Running   0          105m
```

1. Verify that the original DB still has the original data and now the new one:
```
kubectl exec -it $MYSQL_POD -- mysql -uroot -prootpassword -e "USE users_db; SELECT * FROM users;"

mysql: [Warning] Using a password on the command line interface can be insecure.
+----+-------+
| id | name  |
+----+-------+
|  1 | Alice |
+----+-------+
```
1. Reiterate that this allows you to test schema changes, migrations to DBs safely without braking the shared DB.
