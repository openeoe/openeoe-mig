package migration

import (
	"context"
	v1alpha1 "openeoe/openeoe/apis/migration/v1alpha1"
	"openeoe/openeoe/omcplog"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
)

type deploymentInfo struct {
	migSource      *MigrationControllerResource
	resource       v1alpha1.MigrationSource
	newPv          *v1.PersistentVolume
	newPvc         *v1.PersistentVolumeClaim
	targetResource *appsv1.Deployment
	sourceResource *appsv1.Deployment
	description    string
}

func (d *deploymentInfo) migdeploy() error {
	omcplog.V(3).Info("=== deploy migration start : " + d.resource.ResourceName)

	d.targetResource = &appsv1.Deployment{}
	d.sourceResource = &appsv1.Deployment{}

	omcplog.V(3).Info("SourceClient init...")
	omcplog.V(3).Info("TargetClient init...")
	sourceClient := d.migSource.sourceClient
	targetClient := d.migSource.targetClient
	omcplog.V(3).Info("SourceClient init complete")
	omcplog.V(3).Info("TargetClient init complete")

	omcplog.V(3).Info("Make resource info...")

	sourceGetErr := sourceClient.Get(context.TODO(), d.sourceResource, d.migSource.nameSpace, d.resource.ResourceName)
	if sourceGetErr != nil {
		omcplog.Error("get source cluster error : ", sourceGetErr)
		return sourceGetErr
	} else {
		omcplog.V(3).Info("get source info : " + d.resource.ResourceName)
	}
	omcplog.V(3).Info("Make deployment resource info complete")

	d.targetResource = d.sourceResource.DeepCopy() //migLinkShare 함수 실행전 (update)에 해야 링크쉐어 pvc가 포함되지 않는다.
	d.targetResource.ObjectMeta.ResourceVersion = ""
	d.targetResource.Spec.Template.ResourceVersion = ""
	d.targetResource.ResourceVersion = ""
	var linkShareErr error
	linkShareErr = d.migLinkShare()
	if linkShareErr != nil {
		omcplog.Error("run link share error : ", linkShareErr)
		return linkShareErr
	}

	omcplog.V(3).Info("Create for target cluster")
	targetErr := targetClient.Create(context.TODO(), d.targetResource)
	if targetErr != nil {
		if strings.Contains(targetErr.Error(), "already exists") {
			omcplog.V(3).Info("target cluster create error : ", targetErr)
			omcplog.V(3).Info("continue...")
		} else {
			omcplog.Error("target cluster create error : ", targetErr)
			return targetErr
		}
	}
	omcplog.V(3).Info("Create for target cluster end")

	d.description = "[deploy]" + d.resource.ResourceName
	return nil
}

// migdeployWithVolume 함수 내에서 직접 구동되는 linkShare 함수
func (d *deploymentInfo) migLinkShare() error {
	omcplog.V(3).Info("LinkShare init...")
	if !d.migSource.pvCheck {
		omcplog.V(3).Info("not Volume..... link share Skip!!!!")
		return nil
	}
	nameSpace := d.migSource.nameSpace
	sourceClient := d.migSource.sourceClient
	sourceCluster := d.migSource.sourceCluster
	volumePath := d.migSource.volumePath
	serviceName := d.migSource.serviceName
	resourceRequire := d.migSource.resourceRequire
	client := cm.Cluster_kubeClients[sourceCluster]
	restconfig := cm.Cluster_configs[sourceCluster]

	var addpvcErr error
	_, addpvcErr, d.newPv, d.newPvc = CreateLinkShare(sourceClient, d.sourceResource, volumePath, serviceName, resourceRequire)
	//addpvcErr, _ := CreateLinkShare(sourceClient, sourceResource, volumePath, serviceName)
	if addpvcErr != nil {
		omcplog.Error("add pvc error : ", addpvcErr)
		return addpvcErr
	}

	omcplog.V(3).Info(" volumePath : " + volumePath)
	omcplog.V(3).Info(" serviceName : " + serviceName)
	omcplog.V(3).Info("LinkShare init end")

	omcplog.V(3).Info("LinkShare volume copy")
	podName := GetCopyPodName(client, d.sourceResource.Name, nameSpace)
	mkCommand, copyCommand := CopyToNfsCMD(volumePath, serviceName)

	err := LinkShareVolume(client, restconfig, podName, mkCommand, nameSpace)
	if err != nil {
		omcplog.Error("volume make dir error : ", err)
		return err
	} else {
		copyErr := LinkShareVolume(client, restconfig, podName, copyCommand, nameSpace)
		if copyErr != nil {
			omcplog.Error("volume linkshare error : ", copyErr)
			return copyErr
		} else {
			omcplog.V(3).Info("volume linkshare complete")
		}
	}
	omcplog.V(3).Info("LinkShare volume copy end")

	return nil
}

func (d *deploymentInfo) migdeployClose() error {

	omcplog.V(4).Info(d)
	omcplog.V(3).Info("Delete for source cluster")
	nameSpace := d.migSource.nameSpace
	sourceClient := d.migSource.sourceClient
	sourceErr := sourceClient.Delete(context.TODO(), d.sourceResource, nameSpace, d.resource.ResourceName)
	if sourceErr != nil {
		omcplog.Error("source cluster deploy delete error : ", sourceErr)
		return sourceErr
	}

	linkShareCloseErr := d.migLinkShareClose()
	if linkShareCloseErr != nil {
		omcplog.Error("run link share close error : ", linkShareCloseErr)
		return linkShareCloseErr
	}

	omcplog.V(3).Info("Delete for source cluster end")
	return nil
}

// migdeployWithVolume 함수 내에서 직접 구동되는 linkShare 실행 후 close 함수.
func (d *deploymentInfo) migLinkShareClose() error {
	omcplog.V(3).Info("LinkShare close...")
	if !d.migSource.pvCheck {
		omcplog.V(3).Info("not Volume..... link share close Skip!!!!")
		return nil
	}
	nameSpace := d.migSource.nameSpace
	sourceClient := d.migSource.sourceClient
	pvcError := sourceClient.Delete(context.TODO(), d.newPvc, nameSpace, d.newPvc.Name)
	if pvcError != nil {
		omcplog.Error("source cluster pvc delete error : ", pvcError)
		return pvcError
	}
	pvErr := sourceClient.Delete(context.TODO(), d.newPv, nameSpace, d.newPv.Name)
	if pvErr != nil {
		omcplog.Error("source cluster pv delete error: ", pvErr)
		return pvErr
	}
	return nil
}
