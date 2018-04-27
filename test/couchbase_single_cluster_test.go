package test

import (
	"testing"
	"path/filepath"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/gruntwork-io/terratest/modules/test-structure"
	"github.com/gruntwork-io/terratest/modules/random"
)

const couchbaseClusterVarName = "cluster_name"

func TestIntegrationCouchbaseCommunitySingleClusterUbuntu(t *testing.T) {
	t.Parallel()
	testCouchbaseSingleCluster(t, "TestIntegrationCouchbaseCommunitySingleClusterUbuntu", "ubuntu", "community")
}

func TestIntegrationCouchbaseCommunitySingleClusterAmazonLinux(t *testing.T) {
	t.Parallel()
	testCouchbaseSingleCluster(t, "TestIntegrationCouchbaseCommunitySingleClusterAmazonLinux", "amazon-linux", "community")
}

func TestIntegrationCouchbaseEnterpriseSingleClusterUbuntu(t *testing.T) {
	t.Parallel()
	testCouchbaseSingleCluster(t, "TestIntegrationCouchbaseEnterpriseSingleClusterUbuntu", "ubuntu", "enterprise")
}

func TestIntegrationCouchbaseEnterpriseSingleClusterAmazonLinux(t *testing.T) {
	t.Parallel()
	testCouchbaseSingleCluster(t, "TestIntegrationCouchbaseEnterpriseSingleClusterAmazonLinux", "amazon-linux", "enterprise")
}

func testCouchbaseSingleCluster(t *testing.T, testName string, osName string, edition string) {
	examplesFolder := test_structure.CopyTerraformFolderToTemp(t, "../", "examples", testName)
	couchbaseAmiDir := filepath.Join(examplesFolder, "couchbase-ami")
	couchbaseSingleClusterDir := filepath.Join(examplesFolder, "couchbase-single-cluster")

	test_structure.RunTestStage(t, "setup_ami", func() {
		awsRegion := getRandomAwsRegion(t)
		uniqueId := random.UniqueId()

		amiId := buildCouchbaseAmi(t, osName, couchbaseAmiDir, edition, awsRegion, uniqueId)

		test_structure.SaveAmiId(t, couchbaseSingleClusterDir, amiId)
		test_structure.SaveString(t, couchbaseSingleClusterDir, savedAwsRegion, awsRegion)
		test_structure.SaveString(t, couchbaseSingleClusterDir, savedUniqueId, uniqueId)
	})

	test_structure.RunTestStage(t, "setup_deploy", func() {
		amiId := test_structure.LoadAmiId(t, couchbaseSingleClusterDir)
		awsRegion := test_structure.LoadString(t, couchbaseSingleClusterDir, savedAwsRegion)
		uniqueId := test_structure.LoadString(t, couchbaseSingleClusterDir, savedUniqueId)

		terraformOptions := &terraform.Options{
			TerraformDir: couchbaseSingleClusterDir,
			Vars: map[string]interface{}{
				"aws_region":            awsRegion,
				"ami_id":                amiId,
				couchbaseClusterVarName: formatCouchbaseClusterName("single-cluster", uniqueId),
			},
		}

		terraform.Apply(t, terraformOptions)

		test_structure.SaveTerraformOptions(t, couchbaseSingleClusterDir, terraformOptions)
	})

	defer test_structure.RunTestStage(t, "teardown", func() {
		terraformOptions := test_structure.LoadTerraformOptions(t, couchbaseSingleClusterDir)
		terraform.Destroy(t, terraformOptions)
	})

	defer test_structure.RunTestStage(t, "logs", func() {
		terraformOptions := test_structure.LoadTerraformOptions(t, couchbaseSingleClusterDir)
		awsRegion := test_structure.LoadString(t, couchbaseSingleClusterDir, savedAwsRegion)
		testStageLogs(t, terraformOptions, couchbaseClusterVarName, awsRegion)
	})

	test_structure.RunTestStage(t, "validation", func() {
		terraformOptions := test_structure.LoadTerraformOptions(t, couchbaseSingleClusterDir)
		validateSingleClusterWorks(t, terraformOptions, couchbaseClusterVarName, "http")
	})
}
