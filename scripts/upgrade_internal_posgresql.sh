#!/bin/bash

# Usage:
# Used to upgrade on cluster internal Postgresql-10 to Postgresql-13
# ./upgrade_internal_postgresql.sh
# Prerequisite:
# - This script needs to be executed before upgrade to version of 3scale that only supports internal Postgresql-13
# - Logged into the cluster with cluster-admin privileges
# - oc cli


deploymentReady(){
  # Check deploy pod is running and wait till its not
  sleep 40
  DEPLOY_POD_STATUS=$(oc get po --all-namespaces | grep "system-postgresql-.*-deploy" | grep Running | awk '{print $4}')
  DEPLOY_POD_NAME=$(oc get po --all-namespaces | grep "system-postgresql-.*-deploy" | grep Running | awk '{print $2}')
  while [ DEPLOY_POD_STATUS == Running ]
  do
  	DEPLOY_POD_STATUS=$(oc get po --all-namespaces | grep "system-postgresql-.*-deploy" | grep Running | awk '{print $4}')
  	echo the $DEPLOY_POD_NAME deployment is $DEPLOY_POD_STATUS
    POD_NAME=$(oc get po --all-namespaces | grep "system-postgresql-" | awk '{print $2}' | grep -wv deploy)
    POD_STATUS=$(oc get po --all-namespaces | grep $POD_NAME| awk '{print $4}')
    if [ $POD_STATUS == Error ]; then
      echo The pod $POD_NAME status is now in expected $POD_STATUS state, continuing rollout
      break
    fi
  done
}

stopPostgresql(){
  # get the non deployment pod for system-posgresql
  POD=$(oc get po --all-namespaces | grep "system-postgresql-" | awk '{print $2}' | grep -wv deploy)
  POLL=y
  while [[ $POLL == y ]]
  do
    if [[ -z $POD ]]; then
      POLL=y
      POD=$(oc get po --all-namespaces | grep "system-postgresql-" | awk '{print $2}' | grep -wv deploy)
    else
      POLL=n
    fi
  done

  oc exec -it $POD -c system-postgresql -n $THREESCALE_NS -- /usr/bin/pg_ctl stop -D /var/lib/pgsql/data/userdata
  oc exec -it $POD -c system-postgresql -n $THREESCALE_NS  -- rm /var/lib/pgsql/data/userdata/postmaster.pid
}

echo " "
echo "###################################################################################################################"
echo "# Upgrade Postgresql-v10 to Postgresql-v13"
echo "###################################################################################################################"

# Get the 3scale operator csv name and namespace
THREESCALE_OPERATOR_VERSION=$(oc get csv --all-namespaces | grep 3scale-operator | awk '{print $2}')
THREESCALE_NS=$(oc get csv --all-namespaces | grep 3scale-operator | awk '{print $1}')


# Stop postgres and remove the lock file
stopPostgresql

# update the image in the CSV, need some check to see if the deployment is finished
echo " "
echo "###################################################################################################################"
echo "# Starting upgrade to Postgresql-v12"
echo "###################################################################################################################"
oc patch ClusterServiceVersion $THREESCALE_OPERATOR_VERSION -n $THREESCALE_NS --type='json' -p='[{"op": "replace", "path": "/spec/install/spec/deployments/0/spec/template/spec/containers/0/env/9/value", "value":"centos/postgresql-12-centos7"}]'
deploymentReady
oc set env dc/system-postgresql -n $THREESCALE_NS POSTGRESQL_UPGRADE=copy
deploymentReady
oc set env dc/system-postgresql -n $THREESCALE_NS POSTGRESQL_UPGRADE-
deploymentReady
# check the posgresql version in container
oc exec -it $(oc get po --all-namespaces | grep "system-postgresql-" | awk '{print $2}' | grep -wv deploy) -n $THREESCALE_NS -- postgres -V

echo " "
echo "###################################################################################################################"
echo "# Starting upgrade to Postgresql-v13"
echo "###################################################################################################################"
oc patch ClusterServiceVersion $THREESCALE_OPERATOR_VERSION -n $THREESCALE_NS --type='json' -p='[{"op": "replace", "path": "/spec/install/spec/deployments/0/spec/template/spec/containers/0/env/9/value", "value":"centos/postgresql-13-centos7"}]'
deploymentReady
oc set env dc/system-postgresql -n $THREESCALE_NS POSTGRESQL_UPGRADE=copy
deploymentReady
oc set env dc/system-postgresql -n $THREESCALE_NS POSTGRESQL_UPGRADE-
sleep 40
# check the posgresql version in container
oc exec -it $(oc get po --all-namespaces | grep "system-postgresql-" | awk '{print $2}' | grep -wv deploy) -n $THREESCALE_NS -- postgres -V
echo " "
echo "###################################################################################################################"
echo "# Upgrade to Postgresql-v13 finished"
echo "###################################################################################################################"


