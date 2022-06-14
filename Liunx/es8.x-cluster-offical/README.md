> `docker-compose`文件来源于官网文档。

客户端连接es集群以及检查集群健康状态.

```bash
❯ curl --cacert ca.crt https://localhost:9200/_cluster/health/ -u elastic:Elastic123 | jq
  % Total    % Received % Xferd  Average Speed   Time    Time     Time  Current
                                 Dload  Upload   Total   Spent    Left  Speed
100   392  100   392    0     0   9305      0 --:--:-- --:--:-- --:--:-- 10315
{
  "cluster_name": "docker-cluster",
  "status": "green",
  "timed_out": false,
  "number_of_nodes": 3,
  "number_of_data_nodes": 3,
  "active_primary_shards": 16,
  "active_shards": 32,
  "relocating_shards": 0,
  "initializing_shards": 0,
  "unassigned_shards": 0,
  "delayed_unassigned_shards": 0,
  "number_of_pending_tasks": 0,
  "number_of_in_flight_fetch": 0,
  "task_max_waiting_in_queue_millis": 0,
  "active_shards_percent_as_number": 100
}
```

注意请求时，带上`ca.crt`自签证书，以及basic_auth的账户名和密码，账户名和密码都是环境变量，可以在`.env`文件以及`.yml`文件中找到.

`ca.crt`证书的位置(以`kibana`为例)：

```bash
kibana@b4c409efb181:~$ ls /usr/share/kibana/config/certs
ca  ca.zip  certs.zip  es01  es02  es03  instances.yml
kibana@b4c409efb181:~/config/certs$ ls
ca  ca.zip  certs.zip  es01  es02  es03  instances.yml
kibana@b4c409efb181:~/config/certs$ ls ca
ca.crt  ca.key
kibana@b4c409efb181:~/config/certs$ ls es01/
es01.crt  es01.key
```
