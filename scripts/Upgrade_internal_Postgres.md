# Upgrade Internal on Cluster Posgresql v10 to v13

Firstly it is not recommended that you run production workloads on an on-cluster Postgresql. As 3scale operator 
will shortly be removing support for Internal on-cluster Postgresql v10, It's recommend that you migrate from 
Postgresql-v10 to Postgresql-v13.

The following script `upgrade_internal_postgresql.sh` will carry out this upgrade. It has to be run before the 3scale 
operator is upgrade to a version that doesn't support Postgresql-v10

Location of script is in the [3scale-operator](https://github.com/3scale/3scale-operator) git repo in the scripts directory. 
- bash script and needs to be executed on a supporting system(mac/linux). 
- You need to be logged into the cluster with cluster-admin privileges
- oc cli(4 >)



Usage is as follows
```bash
./upgrade_internal_posgresql.sh
 
###################################################################################################################
# Upgrade Postgresql-v10 to Postgresql-v13
###################################################################################################################
waiting for server to shut down....command terminated with exit code 137
 
###################################################################################################################
# Starting upgrade to Postgresql-v12
###################################################################################################################
clusterserviceversion.operators.coreos.com/dev1675239463-3scale-operator.0.0.1 patched
deploymentconfig.apps.openshift.io/system-postgresql updated
deploymentconfig.apps.openshift.io/system-postgresql updated
postgres (PostgreSQL) 12.7
 
###################################################################################################################
# Starting upgrade to Postgresql-v13
###################################################################################################################
clusterserviceversion.operators.coreos.com/dev1675239463-3scale-operator.0.0.1 patched
deploymentconfig.apps.openshift.io/system-postgresql updated
deploymentconfig.apps.openshift.io/system-postgresql updated
postgres (PostgreSQL) 13.3
 
###################################################################################################################
# Upgrade to Postgresql-v13 finished
###################################################################################################################
```


