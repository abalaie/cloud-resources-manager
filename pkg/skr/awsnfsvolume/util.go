package awsnfsvolume

import cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"

func getVolumeName(awsVol *cloudresourcesv1beta1.AwsNfsVolume) string {
	if awsVol.Spec.PersistentVolume != nil &&
		len(awsVol.Spec.PersistentVolume.Name) > 0 {
		return awsVol.Spec.PersistentVolume.Name
	}

	return awsVol.Name
}

func getVolumeLabels(awsVol *cloudresourcesv1beta1.AwsNfsVolume) map[string]string {
	result := map[string]string{}
	if awsVol.Spec.PersistentVolume != nil {
		for k, v := range awsVol.Spec.PersistentVolume.Labels {
			result[k] = v
		}
	}
	result[cloudresourcesv1beta1.LabelCloudManaged] = "true"
	return result
}

func getVolumeAnnotations(awsVol *cloudresourcesv1beta1.AwsNfsVolume) map[string]string {
	if awsVol.Spec.PersistentVolume == nil {
		return nil
	}
	result := map[string]string{}
	for k, v := range awsVol.Spec.PersistentVolume.Annotations {
		result[k] = v
	}
	return result
}

func getVolumeClaimName(awsVol *cloudresourcesv1beta1.AwsNfsVolume) string {
	if awsVol.Spec.PersistentVolumeClaim != nil &&
		len(awsVol.Spec.PersistentVolumeClaim.Name) > 0 {
		return awsVol.Spec.PersistentVolumeClaim.Name
	}

	return awsVol.Name
}

func getVolumeClaimLabels(awsVol *cloudresourcesv1beta1.AwsNfsVolume) map[string]string {
	result := map[string]string{}
	if awsVol.Spec.PersistentVolumeClaim != nil {
		for k, v := range awsVol.Spec.PersistentVolumeClaim.Labels {
			result[k] = v
		}
	}
	result[cloudresourcesv1beta1.LabelCloudManaged] = "true"
	return result
}

func getVolumeClaimAnnotations(awsVol *cloudresourcesv1beta1.AwsNfsVolume) map[string]string {
	if awsVol.Spec.PersistentVolumeClaim == nil {
		return nil
	}
	result := map[string]string{}
	for k, v := range awsVol.Spec.PersistentVolumeClaim.Annotations {
		result[k] = v
	}
	return result
}
