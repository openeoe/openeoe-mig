package migration

import (
	"context"
	v1alpha1 "openeoe/openeoe/apis/migration/v1alpha1"
	"openeoe/openeoe/omcplog"
	"strings"

	corev1 "k8s.io/api/core/v1"
)

type svcInfo struct {
	migSource      *MigrationControllerResource
	resource       v1alpha1.MigrationSource
	targetResource *corev1.Service
	sourceResource *corev1.Service
	description    string
}

func (svc *svcInfo) migservice() error {
	omcplog.V(3).Info("=== service migration start")
	svc.targetResource = &corev1.Service{}
	svc.sourceResource = &corev1.Service{}

	omcplog.V(3).Info("SourceClient init...")
	omcplog.V(3).Info("TargetClient init...")
	sourceClient := svc.migSource.sourceClient
	targetClient := svc.migSource.targetClient
	omcplog.V(3).Info("SourceClient init complete")
	omcplog.V(3).Info("TargetClient init complete")

	omcplog.V(3).Info("Make deployment resource info...")
	nameSpace := svc.migSource.nameSpace

	sourceGetErr := sourceClient.Get(context.TODO(), svc.sourceResource, nameSpace, svc.resource.ResourceName)
	if sourceGetErr != nil {
		omcplog.Error("get source cluster error : ", sourceGetErr)
		return sourceGetErr
	} else {
		omcplog.V(3).Info("get source info : " + svc.resource.ResourceName)
	}
	svc.targetResource = svc.sourceResource.DeepCopy()
	svc.targetResource.ObjectMeta.ResourceVersion = ""
	svc.targetResource.ResourceVersion = ""
	omcplog.V(3).Info("Make deployment resource info complete")

	omcplog.V(3).Info("Create for target cluster")
	targetErr := targetClient.Create(context.TODO(), svc.targetResource)
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

	svc.description = "[svc]" + svc.resource.ResourceName
	return nil
}

func (svc *svcInfo) migserviceClose() error {
	omcplog.V(3).Info("Delete for source cluster")
	nameSpace := svc.migSource.nameSpace
	sourceClient := svc.migSource.sourceClient

	sourceErr := sourceClient.Delete(context.TODO(), svc.sourceResource, nameSpace, svc.resource.ResourceName)
	if sourceErr != nil {
		omcplog.Error("source cluster delete rror : ", sourceErr)
		return sourceErr
	}
	omcplog.V(3).Info("Delete for source cluster end")
	return nil
}
