package migration

import (
	"context"
	v1alpha1 "openeoe/openeoe/apis/migration/v1alpha1"
	"openeoe/openeoe/omcplog"
	"strings"

	corev1 "k8s.io/api/core/v1"
)

type pvInfo struct {
	migSource      *MigrationControllerResource
	resource       v1alpha1.MigrationSource
	targetResource *corev1.PersistentVolume
	sourceResource *corev1.PersistentVolume
	description    string
}

func (pv *pvInfo) migpv() error {
	omcplog.V(3).Info("pv migration")
	pv.targetResource = &corev1.PersistentVolume{}
	pv.sourceResource = &corev1.PersistentVolume{}

	sourceClient := pv.migSource.sourceClient
	targetClient := pv.migSource.targetClient
	nameSpace := pv.migSource.nameSpace
	volumePath := pv.migSource.volumePath
	serviceName := pv.migSource.serviceName

	sourceGetErr := sourceClient.Get(context.TODO(), pv.sourceResource, nameSpace, pv.resource.ResourceName)
	if sourceGetErr != nil {
		omcplog.Error("get source cluster error : ", sourceGetErr)
		return sourceGetErr
	}

	omcplog.V(3).Info("Create for target cluster")
	pv.targetResource = GetLinkSharePv(pv.sourceResource, volumePath, serviceName)
	pv.targetResource.Spec.Capacity = pv.sourceResource.Spec.Capacity
	pv.targetResource.ObjectMeta.ResourceVersion = ""
	pv.targetResource.ResourceVersion = ""
	pv.targetResource.Spec.ClaimRef = nil
	pv.targetResource.Labels = pv.sourceResource.DeepCopy().Labels

	targetErr := targetClient.Create(context.TODO(), pv.targetResource)
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

	pv.description = "[pv]" + pv.resource.ResourceName
	return nil
}

func (pv *pvInfo) migpvClose() error {
	omcplog.V(3).Info("Delete for source cluster")
	nameSpace := pv.migSource.nameSpace
	sourceClient := pv.migSource.sourceClient

	sourceErr := sourceClient.Delete(context.TODO(), pv.sourceResource, nameSpace, pv.resource.ResourceName)
	if sourceErr != nil {
		omcplog.Error("source cluster delete error : ", sourceErr)
		return sourceErr
	}

	omcplog.V(3).Info("Delete for source cluster end")
	return nil
}
