package migration

import (
	"context"
	"fmt"
	v1alpha1 "openeoe/openeoe/apis/migration/v1alpha1"
	resourcev1alpha1 "openeoe/openeoe/apis/resource/v1alpha1"
	"openeoe/openeoe/omcplog"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
)

/*
func (r *reconciler) MakeStatusWithResource(instance *v1alpha1.Migration, migraionSource v1alpha1.MigrationServiceSource, resource v1alpha1.MigrationSource, err error) {
	r.makeStatusRun(instance, migraionSource, resource, "", err)
}

func (r *reconciler) MakeStatusWithMigSource(instance *v1alpha1.Migration, migraionSource v1alpha1.MigrationServiceSource, err error) {
	r.makeStatusRun(instance, migraionSource, v1alpha1.MigrationSource{}, "", err)
}

func (r *reconciler) MakeStatus(instance *v1alpha1.Migration, elapsed string, err error) {
	r.makeStatusRun(instance, v1alpha1.MigrationServiceSource{}, v1alpha1.MigrationSource{}, elapsed, err)
}

func (r *reconciler) makeStatusRun(instance *v1alpha1.Migration, migraionSource v1alpha1.MigrationServiceSource, resource v1alpha1.MigrationSource, elapsedTime string, err error) {

	if elapsedTime == "" {
		elapsedTime = "0"
	}
	instance.Status.ElapsedTime = elapsedTime
	omcplog.V(3).Info("elapsedTime : ", elapsedTime)

	if instance.Status.Status == corev1.ConditionFalse {
		tmp := make(map[string]interface{})
		tmp["SourceCluster"] = migraionSource.SourceCluster
		tmp["TargetCluster"] = migraionSource.TargetCluster
		tmp["ServiceName"] = migraionSource.ServiceName
		tmp["NameSpace"] = migraionSource.NameSpace
		tmp["Description"] = err.Error()

		jsonTmp, err := json.Marshal(tmp)
		if err != nil {
			omcplog.V(3).Info(err, "-----------")
		}
		instance.Status.Description = string(jsonTmp)
	}

	err = r.live.Status().Update(context.TODO(), instance)
	if err != nil {
		omcplog.V(3).Info(err, "-----------")
	}
}
*/

func (r *reconciler) makeStatusRun(instance *v1alpha1.Migration, status corev1.ConditionStatus, description string, isZeroDownTime corev1.ConditionStatus, elapsedTime string, err error) {

	if elapsedTime == "" {
		elapsedTime = "-"
	}

	instance.Status.ElapsedTime = elapsedTime
	instance.Status.Status = status
	instance.Status.Description = description
	instance.Status.IsZeroDownTime = isZeroDownTime
	instance.Status.ConditionProgress = fmt.Sprintf("%f", float64(r.progressCurrent)/float64(r.progressMax)*100) + "%"

	omcplog.V(3).Info(instance.Status)
	omcplog.V(3).Info("progressCurrent : ", r.progressCurrent)
	omcplog.V(3).Info("progressMax : ", r.progressMax)

	omcplog.V(3).Info("elapsedTime : ", instance.Status.ElapsedTime)
	omcplog.V(3).Info("Status : ", instance.Status.Status)
	omcplog.V(3).Info("Description : ", instance.Status.Description)
	omcplog.V(3).Info("isZeroDownTime : ", instance.Status.IsZeroDownTime)
	omcplog.V(3).Info("progressCurrent : ", r.progressCurrent)
	omcplog.V(3).Info("progressMax : ", r.progressMax)
	omcplog.V(3).Info("ConditionProgress : ", instance.Status.ConditionProgress)

	err = r.live.Status().Update(context.TODO(), instance)
	if err != nil {
		omcplog.V(3).Info(err, "-----------")
	}
}

// setBeforeOpenmcpDeploymnet : openeoe deployment 가 있는지 파악하여 없으면 스킵하고, 있으면 openeoe deployment의 CheckSubResource 기능을 키고, r.moveCount 에 값을 넣는 기능.
func (r *reconciler) setBeforeOpenmcpDeploymnet(migraionSource v1alpha1.MigrationServiceSource, idx int) error {
	odeploy := &resourcev1alpha1.OpenEOEDeployment{}
	err := r.live.Get(context.TODO(), types.NamespacedName{Name: migraionSource.MigrationSources[idx].ResourceName, Namespace: migraionSource.NameSpace}, odeploy)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			omcplog.V(4).Info("setBeforeOpenmcpDeploymnet skip : " + migraionSource.MigrationSources[idx].ResourceName)
			return nil
		} else {
			omcplog.Error("setBeforeOpenmcpDeploymnet error : ", err)
		}
		return err
	}
	omcplog.V(4).Info("--- odeploy status ---")
	omcplog.V(4).Info(odeploy.Status)
	moveCount := odeploy.Status.ClusterMaps[migraionSource.SourceCluster]
	omcplog.V(4).Info(moveCount)
	r.moveCount = moveCount

	odeploy.Status.CheckSubResource = true
	err = r.live.Status().Update(context.TODO(), odeploy)
	if err != nil {
		omcplog.V(3).Info(err, "-----------")
	}
	omcplog.V(4).Info("setBeforeOpenmcpDeploymnet end")
	omcplog.V(4).Info("---------------")

	return nil
}

// setAfterOpenmcpDeploymnet : openeoe deployment 가 있는지 파악하여 없으면 스킵하고 있으면 openeoe deployment의 CheckSubResource 기능을 끄고, r.moveCount 의 값을 이용하여 ClusterMaps의 개수를 조정하는 함수.
func (r *reconciler) setAfterOpenmcpDeploymnet(micSource MigrationControllerResource, idx int) error {
	odeploy := &resourcev1alpha1.OpenEOEDeployment{}
	err := r.live.Get(context.TODO(), types.NamespacedName{Name: micSource.resourceList[idx].ResourceName, Namespace: micSource.nameSpace}, odeploy)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			omcplog.V(4).Info("setAfterOpenmcpDeploymnet skip : " + micSource.resourceList[idx].ResourceName)
			return nil
		} else {
			omcplog.Error("setAfterOpenmcpDeploymnet error : ", err)
		}
		return err
	}
	omcplog.V(4).Info("--- odeploy ---")
	omcplog.V(4).Info(odeploy.Status)
	omcplog.V(4).Info(r.moveCount)

	odeploy.Status.CheckSubResource = false
	odeploy.Status.ClusterMaps[micSource.sourceCluster] -= r.moveCount
	odeploy.Status.ClusterMaps[micSource.targetCluster] += r.moveCount

	omcplog.V(4).Info("--- convert odeploy ---")
	omcplog.V(4).Info(odeploy.Status)

	err = r.live.Status().Update(context.TODO(), odeploy)
	if err != nil {
		omcplog.V(3).Info(err, "-----------")
	}
	omcplog.V(4).Info("setAfterOpenmcpDeploymnet end")
	omcplog.V(4).Info("---------------")

	return nil
}
