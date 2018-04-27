package test

import (
	"testing"
	"path/filepath"
	"fmt"
	"github.com/gruntwork-io/terratest/modules/test-structure"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/gruntwork-io/terratest/modules/random"
)

const dataNodeClusterVarName = "couchbase_data_node_cluster_name"
const indexQuerySearchClusterVarName = "couchbase_index_query_search_node_cluster_name"
const syncGatewayClusterVarName = "sync_gateway_cluster_name"

func TestIntegrationCouchbaseEnterpriseMultiClusterUbuntu(t *testing.T) {
	t.Parallel()
	testCouchbaseMultiCluster(t, "TestIntegrationCouchbaseEnterpriseMultiClusterUbuntu", "ubuntu", "enterprise")
}

func TestIntegrationCouchbaseEnterpriseMultiClusterAmazonLinux(t *testing.T) {
	t.Parallel()
	testCouchbaseMultiCluster(t, "TestIntegrationCouchbaseEnterpriseMultiClusterAmazonLinux", "amazon-linux", "enterprise")
}

func testCouchbaseMultiCluster(t *testing.T, testName string, osName string, edition string) {
	examplesFolder := test_structure.CopyTerraformFolderToTemp(t, "../", "examples", testName)
	couchbaseAmiDir := filepath.Join(examplesFolder, "couchbase-ami")
	couchbaseMultiClusterDir := filepath.Join(examplesFolder, "couchbase-multi-cluster")

	test_structure.RunTestStage(t, "setup_ami", func() {
		awsRegion := getRandomAwsRegion(t)
		uniqueId := random.UniqueId()

		amiId := buildCouchbaseAmi(t, osName, couchbaseAmiDir, edition, awsRegion, uniqueId)

		test_structure.SaveAmiId(t, couchbaseMultiClusterDir, amiId)
		test_structure.SaveString(t, couchbaseMultiClusterDir, savedAwsRegion, awsRegion)
		test_structure.SaveString(t, couchbaseMultiClusterDir, savedUniqueId, uniqueId)
	})

	test_structure.RunTestStage(t, "setup_deploy", func() {
		amiId := test_structure.LoadAmiId(t, couchbaseMultiClusterDir)
		awsRegion := test_structure.LoadString(t, couchbaseMultiClusterDir, savedAwsRegion)
		uniqueId := test_structure.LoadString(t, couchbaseMultiClusterDir, savedUniqueId)

		terraformOptions := &terraform.Options{
			TerraformDir: couchbaseMultiClusterDir,
			Vars: map[string]interface{} {
				"aws_region":                   awsRegion,
				"ami_id":                       amiId,
				dataNodeClusterVarName:         formatCouchbaseClusterName("data", uniqueId),
				indexQuerySearchClusterVarName: formatCouchbaseClusterName("search", uniqueId),
				syncGatewayClusterVarName: 		formatCouchbaseClusterName("sync", uniqueId),
			},
		}

		terraform.Apply(t, terraformOptions)

		test_structure.SaveTerraformOptions(t, couchbaseMultiClusterDir, terraformOptions)
	})

	defer test_structure.RunTestStage(t, "teardown", func() {
		terraformOptions := test_structure.LoadTerraformOptions(t, couchbaseMultiClusterDir)
		terraform.Destroy(t, terraformOptions)
	})

	defer test_structure.RunTestStage(t, "logs", func() {
		terraformOptions := test_structure.LoadTerraformOptions(t, couchbaseMultiClusterDir)
		awsRegion := test_structure.LoadString(t, couchbaseMultiClusterDir, savedAwsRegion)

		testStageLogs(t, terraformOptions, dataNodeClusterVarName, awsRegion)
		testStageLogs(t, terraformOptions, indexQuerySearchClusterVarName, awsRegion)
		testStageLogs(t, terraformOptions, syncGatewayClusterVarName, awsRegion)
	})

	test_structure.RunTestStage(t, "validation", func() {
		terraformOptions := test_structure.LoadTerraformOptions(t, couchbaseMultiClusterDir)
		clusterName := getClusterName(t, dataNodeClusterVarName, terraformOptions)

		couchbaseDataNodesUrl := fmt.Sprintf("http://%s:%s@%s", usernameForTest, passwordForTest, terraform.OutputRequired(t, terraformOptions, "couchbase_data_nodes_web_console_url"))
		couchbaseIndexSearchQueryNodesUrl := fmt.Sprintf("http://%s:%s@%s", usernameForTest, passwordForTest, terraform.OutputRequired(t, terraformOptions, "couchbase_index_query_search_nodes_web_console_url"))
		syncGatewayUrl := fmt.Sprintf("http://%s/%s", terraform.OutputRequired(t, terraformOptions, "sync_gateway_url"), clusterName)

		checkCouchbaseConsoleIsRunning(t, couchbaseDataNodesUrl)
		checkCouchbaseClusterIsInitialized(t, couchbaseDataNodesUrl, 5)
		checkCouchbaseDataNodesWorking(t, couchbaseDataNodesUrl)
		checkCouchbaseConsoleIsRunning(t, couchbaseIndexSearchQueryNodesUrl)
		checkSyncGatewayWorking(t, syncGatewayUrl)
	})
}

