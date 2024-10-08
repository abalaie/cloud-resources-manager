package mock

import (
	"context"
	"fmt"
	"sync"

	aws "github.com/aws/aws-sdk-go-v2/aws"
	ec2Types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/elasticache"
	elasticacheTypes "github.com/aws/aws-sdk-go-v2/service/elasticache/types"
	secretsmanager "github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	secretsmanagerTypes "github.com/aws/aws-sdk-go-v2/service/secretsmanager/types"
	"github.com/elliotchance/pie/v2"
	"github.com/google/uuid"
	awsmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/meta"
	awsclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/redisinstance/client"
	"k8s.io/utils/ptr"
)

type AwsElastiCacheMockUtils interface {
	GetAwsElastiCacheByName(name string) *elasticacheTypes.ReplicationGroup
	SetAwsElastiCacheLifeCycleState(name string, state awsmeta.ElastiCacheState)
	SetAwsElastiCacheUserGroupLifeCycleState(name string, state awsmeta.ElastiCacheUserGroupState)
	DeleteAwsElastiCacheByName(name string)
	DeleteAwsElastiCacheUserGroupByName(name string)
	DescribeAwsElastiCacheParametersByName(groupName string) map[string]string
}

func getDefaultParams() map[string]elasticacheTypes.Parameter {
	return map[string]elasticacheTypes.Parameter{
		"maxmemory-policy": {
			ParameterName:  ptr.To("maxmemory-policy"),
			ParameterValue: ptr.To("volatile-lru"),
		},
		"active-defrag-cycle-max": {
			ParameterName:  ptr.To("active-defrag-cycle-max"),
			ParameterValue: ptr.To("75"),
		},
		"acl-pubsub-default": {
			ParameterName:  ptr.To("acl-pubsub-default"),
			ParameterValue: ptr.To("allchannels"),
		},
	}
}

type elastiCacheClientFake struct {
	subnetGroupMutex    *sync.Mutex
	parameterGroupMutex *sync.Mutex
	elasticacheMutex    *sync.Mutex
	secretStoreMutex    *sync.Mutex
	userGroupsMutex     *sync.Mutex
	securityGroupsMutex *sync.Mutex
	replicationGroups   map[string]*elasticacheTypes.ReplicationGroup
	cacheClusters       map[string]*elasticacheTypes.CacheCluster
	parameters          map[string]map[string]elasticacheTypes.Parameter
	parameterGroups     map[string]*elasticacheTypes.CacheParameterGroup
	subnetGroups        map[string]*elasticacheTypes.CacheSubnetGroup
	userGroups          map[string]*elasticacheTypes.UserGroup
	secretStore         map[string]*secretsmanager.GetSecretValueOutput
	securityGroups      []*ec2Types.SecurityGroup
}

func (client *elastiCacheClientFake) GetAwsElastiCacheByName(name string) *elasticacheTypes.ReplicationGroup {
	return client.replicationGroups[name]
}

func (client *elastiCacheClientFake) SetAwsElastiCacheLifeCycleState(name string, state awsmeta.ElastiCacheState) {
	if instance, ok := client.replicationGroups[name]; ok {
		instance.Status = ptr.To(state)
	}
}

func (client *elastiCacheClientFake) SetAwsElastiCacheUserGroupLifeCycleState(name string, state awsmeta.ElastiCacheUserGroupState) {
	if instance, ok := client.userGroups[name]; ok {
		instance.Status = ptr.To(state)
	}
}

func (client *elastiCacheClientFake) DeleteAwsElastiCacheByName(name string) {
	client.elasticacheMutex.Lock()
	defer client.elasticacheMutex.Unlock()

	delete(client.replicationGroups, name)
}

func (client *elastiCacheClientFake) DeleteAwsElastiCacheUserGroupByName(name string) {
	client.userGroupsMutex.Lock()
	defer client.userGroupsMutex.Unlock()

	delete(client.userGroups, name)
}

func (client *elastiCacheClientFake) DescribeAwsElastiCacheParametersByName(groupName string) map[string]string {
	result := map[string]string{}

	for _, parameter := range pie.Values(client.parameters[groupName]) {
		result[*parameter.ParameterName] = *parameter.ParameterValue
	}

	return result
}

func (client *elastiCacheClientFake) DescribeElastiCacheSubnetGroup(ctx context.Context, name string) ([]elasticacheTypes.CacheSubnetGroup, error) {
	client.subnetGroupMutex.Lock()
	defer client.subnetGroupMutex.Unlock()

	subnetGroup := client.subnetGroups[name]

	if subnetGroup == nil {
		return []elasticacheTypes.CacheSubnetGroup{}, nil
	}

	return []elasticacheTypes.CacheSubnetGroup{*subnetGroup}, nil
}

func (client *elastiCacheClientFake) CreateElastiCacheSubnetGroup(ctx context.Context, name string, subnetIds []string, tags []elasticacheTypes.Tag) (*elasticache.CreateCacheSubnetGroupOutput, error) {
	client.subnetGroupMutex.Lock()
	defer client.subnetGroupMutex.Unlock()

	client.subnetGroups[name] = &elasticacheTypes.CacheSubnetGroup{
		CacheSubnetGroupName: ptr.To(name),
	}

	return &elasticache.CreateCacheSubnetGroupOutput{CacheSubnetGroup: &elasticacheTypes.CacheSubnetGroup{
		CacheSubnetGroupName: ptr.To(name),
	}}, nil
}

func (client *elastiCacheClientFake) DeleteElastiCacheSubnetGroup(ctx context.Context, name string) error {
	client.subnetGroupMutex.Lock()
	defer client.subnetGroupMutex.Unlock()

	delete(client.subnetGroups, name)

	return nil
}

func (client *elastiCacheClientFake) DescribeElastiCacheParameterGroup(ctx context.Context, name string) ([]elasticacheTypes.CacheParameterGroup, error) {
	client.parameterGroupMutex.Lock()
	defer client.parameterGroupMutex.Unlock()

	parameterGroup := client.parameterGroups[name]

	if parameterGroup == nil {
		return []elasticacheTypes.CacheParameterGroup{}, nil
	}

	return []elasticacheTypes.CacheParameterGroup{*parameterGroup}, nil
}

func (client *elastiCacheClientFake) CreateElastiCacheParameterGroup(ctx context.Context, name, family string, tags []elasticacheTypes.Tag) (*elasticache.CreateCacheParameterGroupOutput, error) {
	client.parameterGroupMutex.Lock()
	defer client.parameterGroupMutex.Unlock()

	client.parameterGroups[name] = &elasticacheTypes.CacheParameterGroup{
		CacheParameterGroupName:   ptr.To(name),
		CacheParameterGroupFamily: ptr.To(family),
	}

	client.parameters[name] = getDefaultParams()

	return &elasticache.CreateCacheParameterGroupOutput{
		CacheParameterGroup: client.parameterGroups[name],
	}, nil
}

func (client *elastiCacheClientFake) DeleteElastiCacheParameterGroup(ctx context.Context, name string) error {
	client.parameterGroupMutex.Lock()
	defer client.parameterGroupMutex.Unlock()

	delete(client.parameterGroups, name)

	return nil
}

func (client *elastiCacheClientFake) DescribeElastiCacheParameters(ctx context.Context, groupName string) ([]elasticacheTypes.Parameter, error) {
	client.parameterGroupMutex.Lock()
	defer client.parameterGroupMutex.Unlock()

	return pie.Values(client.parameters[groupName]), nil
}

func (client *elastiCacheClientFake) ModifyElastiCacheParameterGroup(ctx context.Context, groupName string, parameters []elasticacheTypes.ParameterNameValue) error {
	client.parameterGroupMutex.Lock()
	defer client.parameterGroupMutex.Unlock()

	for _, parameter := range parameters {
		client.parameters[groupName][ptr.Deref(parameter.ParameterName, "")] = elasticacheTypes.Parameter{ParameterName: parameter.ParameterName, ParameterValue: parameter.ParameterValue}
	}

	return nil
}

func (client *elastiCacheClientFake) DescribeEngineDefaultParameters(ctx context.Context, family string) ([]elasticacheTypes.Parameter, error) {
	return pie.Values(getDefaultParams()), nil
}

func (client *elastiCacheClientFake) GetAuthTokenSecretValue(ctx context.Context, secretName string) (*secretsmanager.GetSecretValueOutput, error) {
	client.secretStoreMutex.Lock()
	defer client.secretStoreMutex.Unlock()

	return client.secretStore[secretName], nil
}

func (client *elastiCacheClientFake) CreateAuthTokenSecret(ctx context.Context, secretName string, tags []secretsmanagerTypes.Tag) error {
	client.secretStoreMutex.Lock()
	defer client.secretStoreMutex.Unlock()

	client.secretStore[secretName] = &secretsmanager.GetSecretValueOutput{
		Name:         ptr.To(secretName),
		SecretString: ptr.To(uuid.NewString()),
	}

	return nil
}

func (client *elastiCacheClientFake) DeleteAuthTokenSecret(ctx context.Context, secretName string) error {
	client.secretStoreMutex.Lock()
	defer client.secretStoreMutex.Unlock()

	delete(client.secretStore, secretName)

	return nil
}

func (client *elastiCacheClientFake) DescribeElastiCacheReplicationGroup(ctx context.Context, clusterId string) ([]elasticacheTypes.ReplicationGroup, error) {
	client.elasticacheMutex.Lock()
	defer client.elasticacheMutex.Unlock()

	cacheCluster := client.replicationGroups[clusterId]

	if cacheCluster == nil {
		return []elasticacheTypes.ReplicationGroup{}, nil
	}

	return []elasticacheTypes.ReplicationGroup{*cacheCluster}, nil
}

func (client *elastiCacheClientFake) CreateElastiCacheReplicationGroup(ctx context.Context, tags []elasticacheTypes.Tag, options awsclient.CreateElastiCacheClusterOptions) (*elasticache.CreateReplicationGroupOutput, error) {
	client.elasticacheMutex.Lock()
	defer client.elasticacheMutex.Unlock()

	client.cacheClusters[options.Name] = &elasticacheTypes.CacheCluster{
		CacheClusterId:             ptr.To(options.Name),
		PreferredMaintenanceWindow: options.PreferredMaintenanceWindow,
	}

	authTokenEnabled := false
	if options.AuthTokenSecretString != nil {
		authTokenEnabled = true
	}

	client.replicationGroups[options.Name] = &elasticacheTypes.ReplicationGroup{
		ReplicationGroupId:       ptr.To(options.Name),
		Status:                   ptr.To("creating"),
		CacheNodeType:            ptr.To(options.CacheNodeType),
		AutoMinorVersionUpgrade:  ptr.To(options.AutoMinorVersionUpgrade),
		TransitEncryptionEnabled: ptr.To(options.TransitEncryptionEnabled),
		AuthTokenEnabled:         ptr.To(authTokenEnabled),
		MemberClusters:           []string{options.Name},
		UserGroupIds:             []string{},
		NodeGroups: []elasticacheTypes.NodeGroup{
			{
				PrimaryEndpoint: &elasticacheTypes.Endpoint{
					Address: ptr.To("192.168.3.3"),
					Port:    aws.Int32(6949),
				},
				ReaderEndpoint: &elasticacheTypes.Endpoint{
					Address: ptr.To("192.168.3.4"),
					Port:    aws.Int32(6949),
				},
			},
		},
	}
	if options.TransitEncryptionEnabled {
		client.replicationGroups[options.Name].TransitEncryptionMode = elasticacheTypes.TransitEncryptionModeRequired
	}

	return &elasticache.CreateReplicationGroupOutput{}, nil
}

func (client *elastiCacheClientFake) ModifyElastiCacheReplicationGroup(ctx context.Context, id string, options awsclient.ModifyElastiCacheClusterOptions) (*elasticache.ModifyReplicationGroupOutput, error) {
	client.elasticacheMutex.Lock()
	defer client.elasticacheMutex.Unlock()

	if instance, ok := client.replicationGroups[id]; ok {
		instance.Status = ptr.To("modifying")
		if options.CacheNodeType != nil {
			instance.CacheNodeType = options.CacheNodeType
		}
		if options.AutoMinorVersionUpgrade != nil {
			instance.AutoMinorVersionUpgrade = options.AutoMinorVersionUpgrade
		}
		if options.TransitEncryptionEnabled != nil {
			instance.TransitEncryptionEnabled = options.TransitEncryptionEnabled
			if *options.TransitEncryptionEnabled {
				instance.TransitEncryptionMode = elasticacheTypes.TransitEncryptionModeRequired
			} else {
				instance.TransitEncryptionMode = ""
			}
		}
		if options.TransitEncryptionMode != nil {
			instance.TransitEncryptionMode = *options.TransitEncryptionMode
		}

		if len(options.UserGroupIdsToAdd) > 0 {
			instance.UserGroupIds = append(instance.UserGroupIds, options.UserGroupIdsToAdd...)
			instance.AuthTokenEnabled = ptr.To(false)
		}
		if len(options.UserGroupIdsToRemove) > 0 {
			_, remaining := pie.Diff(instance.UserGroupIds, options.UserGroupIdsToRemove)
			instance.UserGroupIds = remaining
		}

		if options.AuthTokenSecretString != nil {
			instance.AuthTokenEnabled = ptr.To(true)
		}
	}

	return &elasticache.ModifyReplicationGroupOutput{}, nil
}

func (client *elastiCacheClientFake) DeleteElastiCacheReplicationGroup(ctx context.Context, id string) error {
	client.elasticacheMutex.Lock()
	defer client.elasticacheMutex.Unlock()

	if instance, ok := client.replicationGroups[id]; ok {
		instance.Status = ptr.To("deleting")
	}

	return nil
}

func (client *elastiCacheClientFake) DescribeElastiCacheCluster(ctx context.Context, id string) ([]elasticacheTypes.CacheCluster, error) {
	client.elasticacheMutex.Lock()
	defer client.elasticacheMutex.Unlock()

	cacheCluster := client.cacheClusters[id]

	if cacheCluster == nil {
		return []elasticacheTypes.CacheCluster{}, nil
	}

	return []elasticacheTypes.CacheCluster{*cacheCluster}, nil
}

func (client *elastiCacheClientFake) DescribeUserGroup(ctx context.Context, id string) (*elasticacheTypes.UserGroup, error) {
	client.userGroupsMutex.Lock()
	defer client.userGroupsMutex.Unlock()

	userGroup := client.userGroups[id]

	return userGroup, nil
}

func (client *elastiCacheClientFake) CreateUserGroup(ctx context.Context, id string, tags []elasticacheTypes.Tag) (*elasticache.CreateUserGroupOutput, error) {
	client.userGroupsMutex.Lock()
	defer client.userGroupsMutex.Unlock()

	client.userGroups[id] = &elasticacheTypes.UserGroup{
		Engine:      ptr.To("redis"),
		UserGroupId: ptr.To(id),
		Status:      ptr.To("creating"),
		UserIds:     []string{"default"},
	}

	return &elasticache.CreateUserGroupOutput{UserGroupId: ptr.To(id)}, nil
}

func (client *elastiCacheClientFake) DeleteUserGroup(ctx context.Context, id string) error {
	client.userGroupsMutex.Lock()
	defer client.userGroupsMutex.Unlock()

	if instance, ok := client.userGroups[id]; ok {
		instance.Status = ptr.To("deleting")
	}

	return nil
}

func (client *elastiCacheClientFake) DescribeElastiCacheSecurityGroups(ctx context.Context, filters []ec2Types.Filter, groupIds []string) ([]ec2Types.SecurityGroup, error) {
	client.securityGroupsMutex.Lock()
	defer client.securityGroupsMutex.Unlock()

	list := append([]*ec2Types.SecurityGroup{}, client.securityGroups...)
	if groupIds != nil {
		list = pie.Filter(list, func(sg *ec2Types.SecurityGroup) bool {
			return pie.Contains(groupIds, ptr.Deref(sg.GroupId, ""))
		})
	}
	if filters != nil {
		list = pie.Filter(list, func(sg *ec2Types.SecurityGroup) bool {
			return anyFilterMatchTags(sg.Tags, filters)
		})
	}
	result := make([]ec2Types.SecurityGroup, 0, len(list))
	for _, x := range list {
		result = append(result, *x)
	}
	return result, nil
}

func (client *elastiCacheClientFake) CreateElastiCacheSecurityGroup(ctx context.Context, vpcId string, name string, tags []ec2Types.Tag) (string, error) {
	client.securityGroupsMutex.Lock()
	defer client.securityGroupsMutex.Unlock()

	tags = append(tags, ec2Types.Tag{
		Key:   ptr.To("vpc-id"),
		Value: ptr.To(vpcId),
	})
	sg := &ec2Types.SecurityGroup{
		Description: ptr.To(name),
		GroupId:     ptr.To(uuid.NewString()),
		GroupName:   ptr.To(name),
		Tags:        tags,
		VpcId:       ptr.To(vpcId),
	}
	client.securityGroups = append(client.securityGroups, sg)
	return ptr.Deref(sg.GroupId, ""), nil
}

func (client *elastiCacheClientFake) AuthorizeElastiCacheSecurityGroupIngress(ctx context.Context, groupId string, ipPermissions []ec2Types.IpPermission) error {
	client.securityGroupsMutex.Lock()
	defer client.securityGroupsMutex.Unlock()

	var securityGroup *ec2Types.SecurityGroup
	for _, sg := range client.securityGroups {
		if ptr.Deref(sg.GroupId, "") == groupId {
			securityGroup = sg
			break
		}
	}
	if securityGroup == nil {
		return fmt.Errorf("security group with id %s does not exist", groupId)
	}
	securityGroup.IpPermissions = ipPermissions
	return nil
}

func (client *elastiCacheClientFake) DeleteElastiCacheSecurityGroup(ctx context.Context, id string) error {
	client.securityGroupsMutex.Lock()
	defer client.securityGroupsMutex.Unlock()

	client.securityGroups = pie.Filter(client.securityGroups, func(sg *ec2Types.SecurityGroup) bool {
		return ptr.Deref(sg.GroupId, "") != id
	})
	return nil
}
