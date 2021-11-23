package migration

import (
	"context"
	v1alpha1 "openeoe/openeoe/apis/migration/v1alpha1"
	"openeoe/openeoe/omcplog"
	"strings"

	corev1 "k8s.io/api/core/v1"
)

type pvcInfo struct {
	migSource      *MigrationControllerResource
	resource       v1alpha1.MigrationSource
	targetResource *corev1.PersistentVolumeClaim
	sourceResource *corev1.PersistentVolumeClaim
	description    string
}

func (pvc *pvcInfo) migpvc() error {
	omcplog.V(3).Info("pvc migration")
	pvc.targetResource = &corev1.PersistentVolumeClaim{}
	pvc.sourceResource = &corev1.PersistentVolumeClaim{}

	sourceClient := pvc.migSource.sourceClient
	targetClient := pvc.migSource.targetClient
	nameSpace := pvc.migSource.nameSpace
	volumePath := pvc.migSource.volumePath
	serviceName := pvc.migSource.serviceName

	// targetGetErr := targetClient.Get(context.TODO(), targetResource, nameSpace, resource.ResourceName)
	// if targetGetErr != nil {
	// 	omcplog.V(3).Info("get target cluster")
	// }
	sourceGetErr := sourceClient.Get(context.TODO(), pvc.sourceResource, nameSpace, pvc.resource.ResourceName)
	if sourceGetErr != nil {
		omcplog.Error("get source cluster error : ", sourceGetErr)
		return sourceGetErr
	}

	omcplog.V(3).Info("Create for target cluster")
	//targetResource = sourceResource
	pvc.targetResource = GetLinkSharePvc(pvc.sourceResource, volumePath, serviceName)
	pvc.targetResource.ObjectMeta.ResourceVersion = ""
	pvc.targetResource.ResourceVersion = ""

	targetErr := targetClient.Create(context.TODO(), pvc.targetResource)
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

	pvc.description = "[pvc]" + pvc.resource.ResourceName
	return nil
}

func (pvc *pvcInfo) migpvcClose() error {
	omcplog.V(3).Info("Delete for source cluster")
	nameSpace := pvc.migSource.nameSpace
	sourceClient := pvc.migSource.sourceClient

	sourceErr := sourceClient.Delete(context.TODO(), pvc.sourceResource, nameSpace, pvc.resource.ResourceName)
	if sourceErr != nil {
		omcplog.Error("source cluster delete error : ", sourceErr)
		return sourceErr
	}

	omcplog.V(3).Info("Delete for source cluster end")
	return nil
}
