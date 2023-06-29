## HDFS-Plugin
> BSc Thesis Supervisor: [Dr. Mahmoud Momtazpour](https://scholar.google.co.za/citations?user=uwozfWkAAAAJ&hl=en)

kubectl plugin for [HDFS-operator](https://github.com/AmirAllahveran/HDFS-operator) project

### How to build
```bash
go build -o kubectl-hdfs main.go
```

### How to install
```bash
cp ./kubectl-hdfs /usr/local/bin
```

### How to use
```bash
kubectl hdfs [cluster-name] "hadoop fs -mkdir /input" --namespace [NAMESPACE]
```
