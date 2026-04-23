package main

import (
	"fmt"
	"strings"

	"github.com/numberly/opentofu-provider-flashblade/pulumi/sdk/go/flashblade"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

// FlashBlade S3 Bucket — Dual-Array Bidirectional Replication (Go)
//
// Mirrors the terraform-flashblade-s3-bucket module:
//   - Object store accounts on both arrays
//   - S3 export policies + account exports
//   - IAM-style access policies (global + per-user)
//   - Named S3 users with per-user policies and access keys
//   - Versioned buckets with quotas
//   - Bidirectional replication via remote credentials + replica links
//   - Lifecycle rules, audit filters, and QoS policy (optional)
//
// Prerequisites:
//   - Array connection configured between both arrays
//   - Servers pre-provisioned on both arrays

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		cfg := config.New(ctx, "")

		// ------------------------------------------------------------------
		// Config
		// ------------------------------------------------------------------
		par5Endpoint := cfg.Require("par5Endpoint")
		pa7Endpoint := cfg.Require("pa7Endpoint")
		par5Token := cfg.RequireSecret("par5ApiToken")
		pa7Token := cfg.RequireSecret("pa7ApiToken")

		par5ArrayName := cfg.Require("par5ArrayName")
		pa7ArrayName := cfg.Require("pa7ArrayName")
		par5ServerName := cfg.Require("par5ServerName")
		pa7ServerName := cfg.Require("pa7ServerName")

		accountName := cfg.Require("accountName")
		bucketName := cfg.Require("bucketName")
		bucketQuotaBytes := cfg.GetInt("bucketQuotaBytes")
		fqdn := cfg.Get("fqdn")

		enableS3TargetReplication := cfg.GetBool("enableS3TargetReplication")
		flashbladeTargetPar5SeesPa7 := cfg.Get("flashbladeTargetPar5SeesPa7")
		flashbladeTargetPa7SeesPar5 := cfg.Get("flashbladeTargetPa7SeesPar5")

		auditEnabled := cfg.GetBool("auditEnabled")
		var auditPrefixes []string
		cfg.GetObject("auditPrefixes", &auditPrefixes)

		var users map[string]struct {
			VaultSecretPath string   `json:"vault_secret_path"`
			S3Actions       []string `json:"s3_actions"`
			S3Resources     []string `json:"s3_resources"`
			Effect          string   `json:"effect"`
			FullAccess      bool     `json:"full_access"`
		}
		cfg.GetObject("users", &users)

		var lifecycleRules map[string]struct {
			Prefix                           string `json:"prefix"`
			Enabled                          bool   `json:"enabled"`
			KeepPreviousVersionFor           *int   `json:"keep_previous_version_for"`
			KeepCurrentVersionFor            *int   `json:"keep_current_version_for"`
			KeepCurrentVersionUntil          *int   `json:"keep_current_version_until"`
			AbortIncompleteMultipartUploadsAfter *int `json:"abort_incomplete_multipart_uploads_after"`
		}
		if err := cfg.GetObject("lifecycleRules", &lifecycleRules); err != nil || len(lifecycleRules) == 0 {
			lifecycleRules = map[string]struct {
				Prefix                           string `json:"prefix"`
				Enabled                          bool   `json:"enabled"`
				KeepPreviousVersionFor           *int   `json:"keep_previous_version_for"`
				KeepCurrentVersionFor            *int   `json:"keep_current_version_for"`
				KeepCurrentVersionUntil          *int   `json:"keep_current_version_until"`
				AbortIncompleteMultipartUploadsAfter *int `json:"abort_incomplete_multipart_uploads_after"`
			}{
				"default": {
					Prefix:                 "",
					Enabled:                true,
					KeepPreviousVersionFor: intPtr(2592000000),
					KeepCurrentVersionFor:  intPtr(2592000000),
					AbortIncompleteMultipartUploadsAfter: intPtr(604800000),
				},
			}
		}

		var qos *struct {
			MaxTotalBytesPerSec *int `json:"max_total_bytes_per_sec"`
			MaxTotalOpsPerSec   *int `json:"max_total_ops_per_sec"`
		}
		cfg.GetObject("qos", &qos)

		// ------------------------------------------------------------------
		// Providers
		// ------------------------------------------------------------------
		providerPar5, err := flashblade.NewProvider(ctx, "par5", &flashblade.ProviderArgs{
			Endpoint: pulumi.String(par5Endpoint),
			Auth: &flashblade.ProviderAuthArgs{
				ApiToken: par5Token,
			},
		})
		if err != nil {
			return err
		}

		providerPa7, err := flashblade.NewProvider(ctx, "pa7", &flashblade.ProviderArgs{
			Endpoint: pulumi.String(pa7Endpoint),
			Auth: &flashblade.ProviderAuthArgs{
				ApiToken: pa7Token,
			},
		})
		if err != nil {
			return err
		}

		// ------------------------------------------------------------------
		// Step 1: Array connections (data sources)
		// ------------------------------------------------------------------
		par5SeesPa7, err := flashblade.LookupArrayConnection(ctx, &flashblade.LookupArrayConnectionArgs{
			RemoteName: pa7ArrayName,
		}, pulumi.Provider(providerPar5))
		if err != nil {
			return err
		}

		pa7SeesPar5, err := flashblade.LookupArrayConnection(ctx, &flashblade.LookupArrayConnectionArgs{
			RemoteName: par5ArrayName,
		}, pulumi.Provider(providerPa7))
		if err != nil {
			return err
		}

		// ------------------------------------------------------------------
		// Step 2: Servers
		// ------------------------------------------------------------------
		par5Server, err := flashblade.LookupServer(ctx, &flashblade.LookupServerArgs{
			Name: par5ServerName,
		}, pulumi.Provider(providerPar5))
		if err != nil {
			return err
		}

		pa7Server, err := flashblade.LookupServer(ctx, &flashblade.LookupServerArgs{
			Name: pa7ServerName,
		}, pulumi.Provider(providerPa7))
		if err != nil {
			return err
		}

		// ------------------------------------------------------------------
		// Step 3: Object store accounts
		// ------------------------------------------------------------------
		accountPar5, err := flashblade.NewObjectStoreAccount(ctx, "par5", &flashblade.ObjectStoreAccountArgs{
			Name:             pulumi.String(accountName),
			SkipDefaultExport: pulumi.Bool(true),
		}, pulumi.Provider(providerPar5))
		if err != nil {
			return err
		}

		accountPa7, err := flashblade.NewObjectStoreAccount(ctx, "pa7", &flashblade.ObjectStoreAccountArgs{
			Name:             pulumi.String(accountName),
			SkipDefaultExport: pulumi.Bool(true),
		}, pulumi.Provider(providerPa7))
		if err != nil {
			return err
		}

		// ------------------------------------------------------------------
		// Step 4: S3 export policies
		// ------------------------------------------------------------------
		s3PolicyPar5, err := flashblade.NewS3ExportPolicy(ctx, "par5", &flashblade.S3ExportPolicyArgs{
			Name:    pulumi.String(fmt.Sprintf("%s-s3-export", accountName)),
			Enabled: pulumi.Bool(true),
		}, pulumi.Provider(providerPar5))
		if err != nil {
			return err
		}

		_, err = flashblade.NewS3ExportPolicyRule(ctx, "par5", &flashblade.S3ExportPolicyRuleArgs{
			PolicyName: s3PolicyPar5.Name,
			Name:       pulumi.String("allows3"),
			Actions:    pulumi.ToStringArray([]string{"pure:S3Access"}),
			Effect:     pulumi.String("allow"),
			Resources:  pulumi.ToStringArray([]string{"*"}),
		}, pulumi.Provider(providerPar5))
		if err != nil {
			return err
		}

		s3PolicyPa7, err := flashblade.NewS3ExportPolicy(ctx, "pa7", &flashblade.S3ExportPolicyArgs{
			Name:    pulumi.String(fmt.Sprintf("%s-s3-export", accountName)),
			Enabled: pulumi.Bool(true),
		}, pulumi.Provider(providerPa7))
		if err != nil {
			return err
		}

		_, err = flashblade.NewS3ExportPolicyRule(ctx, "pa7", &flashblade.S3ExportPolicyRuleArgs{
			PolicyName: s3PolicyPa7.Name,
			Name:       pulumi.String("allows3"),
			Actions:    pulumi.ToStringArray([]string{"pure:S3Access"}),
			Effect:     pulumi.String("allow"),
			Resources:  pulumi.ToStringArray([]string{"*"}),
		}, pulumi.Provider(providerPa7))
		if err != nil {
			return err
		}

		// ------------------------------------------------------------------
		// Step 5: Account exports
		// ------------------------------------------------------------------
		_, err = flashblade.NewObjectStoreAccountExport(ctx, "par5", &flashblade.ObjectStoreAccountExportArgs{
			AccountName: accountPar5.Name,
			ServerName:  pulumi.String(par5Server.Name),
			PolicyName:  s3PolicyPar5.Name,
			Enabled:     pulumi.Bool(true),
		}, pulumi.Provider(providerPar5))
		if err != nil {
			return err
		}

		_, err = flashblade.NewObjectStoreAccountExport(ctx, "pa7", &flashblade.ObjectStoreAccountExportArgs{
			AccountName: accountPa7.Name,
			ServerName:  pulumi.String(pa7Server.Name),
			PolicyName:  s3PolicyPa7.Name,
			Enabled:     pulumi.Bool(true),
		}, pulumi.Provider(providerPa7))
		if err != nil {
			return err
		}

		// ------------------------------------------------------------------
		// Step 6: Access policies (global rw)
		// ------------------------------------------------------------------
		accessPolicyPar5, err := flashblade.NewObjectStoreAccessPolicy(ctx, "par5", &flashblade.ObjectStoreAccessPolicyArgs{
			Name: pulumi.String(fmt.Sprintf("%s/rw", accountName)),
		}, pulumi.Provider(providerPar5))
		if err != nil {
			return err
		}

		_, err = flashblade.NewObjectStoreAccessPolicyRule(ctx, "par5", &flashblade.ObjectStoreAccessPolicyRuleArgs{
			PolicyName: accessPolicyPar5.Name,
			Name:       pulumi.String("allowrw"),
			Effect:     pulumi.String("allow"),
			Actions:    pulumi.ToStringArray([]string{"s3:*"}),
			Resources:  pulumi.ToStringArray([]string{"*"}),
		}, pulumi.Provider(providerPar5))
		if err != nil {
			return err
		}

		accessPolicyPa7, err := flashblade.NewObjectStoreAccessPolicy(ctx, "pa7", &flashblade.ObjectStoreAccessPolicyArgs{
			Name: pulumi.String(fmt.Sprintf("%s/rw", accountName)),
		}, pulumi.Provider(providerPa7))
		if err != nil {
			return err
		}

		_, err = flashblade.NewObjectStoreAccessPolicyRule(ctx, "pa7", &flashblade.ObjectStoreAccessPolicyRuleArgs{
			PolicyName: accessPolicyPa7.Name,
			Name:       pulumi.String("allowrw"),
			Effect:     pulumi.String("allow"),
			Actions:    pulumi.ToStringArray([]string{"s3:*"}),
			Resources:  pulumi.ToStringArray([]string{"*"}),
		}, pulumi.Provider(providerPa7))
		if err != nil {
			return err
		}

		// ------------------------------------------------------------------
		// Step 7: Buckets
		// ------------------------------------------------------------------
		var quotaLimit *int
		if bucketQuotaBytes != 0 {
			quotaLimit = &bucketQuotaBytes
		}
		hardLimit := quotaLimit != nil

		bucketPar5, err := flashblade.NewBucket(ctx, "par5", &flashblade.BucketArgs{
			Name:                    pulumi.String(bucketName),
			Account:                 accountPar5.Name,
			Versioning:              pulumi.String("enabled"),
			QuotaLimit:              pulumi.IntPtrFromPtr(quotaLimit),
			HardLimitEnabled:        pulumi.Bool(hardLimit),
			DestroyEradicateOnDelete: pulumi.Bool(false),
		}, pulumi.Provider(providerPar5), pulumi.Timeouts(&pulumi.CustomTimeouts{
			Create: "20m",
			Update: "20m",
			Delete: "30m",
		}))
		if err != nil {
			return err
		}

		bucketPa7, err := flashblade.NewBucket(ctx, "pa7", &flashblade.BucketArgs{
			Name:                    pulumi.String(bucketName),
			Account:                 accountPa7.Name,
			Versioning:              pulumi.String("enabled"),
			QuotaLimit:              pulumi.IntPtrFromPtr(quotaLimit),
			HardLimitEnabled:        pulumi.Bool(hardLimit),
			DestroyEradicateOnDelete: pulumi.Bool(false),
		}, pulumi.Provider(providerPa7), pulumi.Timeouts(&pulumi.CustomTimeouts{
			Create: "20m",
			Update: "20m",
			Delete: "30m",
		}))
		if err != nil {
			return err
		}

		// ------------------------------------------------------------------
		// Step 8: Replication user, policy, and keys
		// ------------------------------------------------------------------
		replicationUserPar5, err := flashblade.NewObjectStoreUser(ctx, "replicationPar5", &flashblade.ObjectStoreUserArgs{
			Name:       pulumi.String(fmt.Sprintf("%s/replication", accountName)),
			FullAccess: pulumi.Bool(true),
		}, pulumi.Provider(providerPar5))
		if err != nil {
			return err
		}

		replicationUserPa7, err := flashblade.NewObjectStoreUser(ctx, "replicationPa7", &flashblade.ObjectStoreUserArgs{
			Name:       pulumi.String(fmt.Sprintf("%s/replication", accountName)),
			FullAccess: pulumi.Bool(true),
		}, pulumi.Provider(providerPa7))
		if err != nil {
			return err
		}

		replicationPolicyPar5, err := flashblade.NewObjectStoreAccessPolicy(ctx, "replicationPar5", &flashblade.ObjectStoreAccessPolicyArgs{
			Name: pulumi.String(fmt.Sprintf("%s/replication", accountName)),
		}, pulumi.Provider(providerPar5))
		if err != nil {
			return err
		}

		_, err = flashblade.NewObjectStoreAccessPolicyRule(ctx, "replicationPar5", &flashblade.ObjectStoreAccessPolicyRuleArgs{
			PolicyName: replicationPolicyPar5.Name,
			Name:       pulumi.String("replicationrw"),
			Effect:     pulumi.String("allow"),
			Actions:    pulumi.ToStringArray([]string{"s3:*"}),
			Resources:  pulumi.ToStringArray([]string{"*"}),
		}, pulumi.Provider(providerPar5))
		if err != nil {
			return err
		}

		replicationPolicyPa7, err := flashblade.NewObjectStoreAccessPolicy(ctx, "replicationPa7", &flashblade.ObjectStoreAccessPolicyArgs{
			Name: pulumi.String(fmt.Sprintf("%s/replication", accountName)),
		}, pulumi.Provider(providerPa7))
		if err != nil {
			return err
		}

		_, err = flashblade.NewObjectStoreAccessPolicyRule(ctx, "replicationPa7", &flashblade.ObjectStoreAccessPolicyRuleArgs{
			PolicyName: replicationPolicyPa7.Name,
			Name:       pulumi.String("replicationrw"),
			Effect:     pulumi.String("allow"),
			Actions:    pulumi.ToStringArray([]string{"s3:*"}),
			Resources:  pulumi.ToStringArray([]string{"*"}),
		}, pulumi.Provider(providerPa7))
		if err != nil {
			return err
		}

		_, err = flashblade.NewObjectStoreUserPolicy(ctx, "replicationPar5", &flashblade.ObjectStoreUserPolicyArgs{
			UserName:   replicationUserPar5.Name,
			PolicyName: replicationPolicyPar5.Name,
		}, pulumi.Provider(providerPar5))
		if err != nil {
			return err
		}

		_, err = flashblade.NewObjectStoreUserPolicy(ctx, "replicationPa7", &flashblade.ObjectStoreUserPolicyArgs{
			UserName:   replicationUserPa7.Name,
			PolicyName: replicationPolicyPa7.Name,
		}, pulumi.Provider(providerPa7))
		if err != nil {
			return err
		}

		replicationKeyPar5, err := flashblade.NewObjectStoreAccessKey(ctx, "par5", &flashblade.ObjectStoreAccessKeyArgs{
			ObjectStoreAccount: accountPar5.Name,
			User:               replicationUserPar5.Name,
		}, pulumi.Provider(providerPar5))
		if err != nil {
			return err
		}

		replicationKeyPa7, err := flashblade.NewObjectStoreAccessKey(ctx, "pa7", &flashblade.ObjectStoreAccessKeyArgs{
			ObjectStoreAccount: accountPa7.Name,
			User:               replicationUserPa7.Name,
			Name:               replicationKeyPar5.Name,
			SecretAccessKey:    replicationKeyPar5.SecretAccessKey,
		}, pulumi.Provider(providerPa7))
		if err != nil {
			return err
		}

		// ------------------------------------------------------------------
		// Step 9: Named S3 users with per-user policies
		// ------------------------------------------------------------------
		userKeysPar5 := make(map[string]*flashblade.ObjectStoreAccessKey)

		for username, userCfg := range users {
			safeName := strings.ReplaceAll(username, "-", "_")

			// Users
			userPar5, err := flashblade.NewObjectStoreUser(ctx, fmt.Sprintf("userPar5_%s", safeName), &flashblade.ObjectStoreUserArgs{
				Name:       pulumi.String(fmt.Sprintf("%s/%s", accountName, username)),
				FullAccess: pulumi.Bool(userCfg.FullAccess),
			}, pulumi.Provider(providerPar5))
			if err != nil {
				return err
			}

			userPa7, err := flashblade.NewObjectStoreUser(ctx, fmt.Sprintf("userPa7_%s", safeName), &flashblade.ObjectStoreUserArgs{
				Name:       pulumi.String(fmt.Sprintf("%s/%s", accountName, username)),
				FullAccess: pulumi.Bool(userCfg.FullAccess),
			}, pulumi.Provider(providerPa7))
			if err != nil {
				return err
			}

			// Per-user policies
			userPolicyPar5, err := flashblade.NewObjectStoreAccessPolicy(ctx, fmt.Sprintf("userPar5_%s", safeName), &flashblade.ObjectStoreAccessPolicyArgs{
				Name: pulumi.String(fmt.Sprintf("%s/%s", accountName, username)),
			}, pulumi.Provider(providerPar5))
			if err != nil {
				return err
			}

			_, err = flashblade.NewObjectStoreAccessPolicyRule(ctx, fmt.Sprintf("userPar5_%s", safeName), &flashblade.ObjectStoreAccessPolicyRuleArgs{
				PolicyName: userPolicyPar5.Name,
				Name:       pulumi.String(fmt.Sprintf("%srule", strings.ReplaceAll(username, "-", ""))),
				Effect:     pulumi.String(userCfg.Effect),
				Actions:    pulumi.ToStringArray(userCfg.S3Actions),
				Resources:  pulumi.ToStringArray(userCfg.S3Resources),
			}, pulumi.Provider(providerPar5))
			if err != nil {
				return err
			}

			userPolicyPa7, err := flashblade.NewObjectStoreAccessPolicy(ctx, fmt.Sprintf("userPa7_%s", safeName), &flashblade.ObjectStoreAccessPolicyArgs{
				Name: pulumi.String(fmt.Sprintf("%s/%s", accountName, username)),
			}, pulumi.Provider(providerPa7))
			if err != nil {
				return err
			}

			_, err = flashblade.NewObjectStoreAccessPolicyRule(ctx, fmt.Sprintf("userPa7_%s", safeName), &flashblade.ObjectStoreAccessPolicyRuleArgs{
				PolicyName: userPolicyPa7.Name,
				Name:       pulumi.String(fmt.Sprintf("%srule", strings.ReplaceAll(username, "-", ""))),
				Effect:     pulumi.String(userCfg.Effect),
				Actions:    pulumi.ToStringArray(userCfg.S3Actions),
				Resources:  pulumi.ToStringArray(userCfg.S3Resources),
			}, pulumi.Provider(providerPa7))
			if err != nil {
				return err
			}

			// User-to-policy associations
			_, err = flashblade.NewObjectStoreUserPolicy(ctx, fmt.Sprintf("userPar5_%s", safeName), &flashblade.ObjectStoreUserPolicyArgs{
				UserName:   userPar5.Name,
				PolicyName: userPolicyPar5.Name,
			}, pulumi.Provider(providerPar5))
			if err != nil {
				return err
			}

			_, err = flashblade.NewObjectStoreUserPolicy(ctx, fmt.Sprintf("userPa7_%s", safeName), &flashblade.ObjectStoreUserPolicyArgs{
				UserName:   userPa7.Name,
				PolicyName: userPolicyPa7.Name,
			}, pulumi.Provider(providerPa7))
			if err != nil {
				return err
			}

			// Per-user access keys
			keyPar5, err := flashblade.NewObjectStoreAccessKey(ctx, fmt.Sprintf("userPar5_%s", safeName), &flashblade.ObjectStoreAccessKeyArgs{
				ObjectStoreAccount: accountPar5.Name,
				User:               userPar5.Name,
			}, pulumi.Provider(providerPar5))
			if err != nil {
				return err
			}

			_, err = flashblade.NewObjectStoreAccessKey(ctx, fmt.Sprintf("userPa7_%s", safeName), &flashblade.ObjectStoreAccessKeyArgs{
				ObjectStoreAccount: accountPa7.Name,
				User:               userPa7.Name,
				Name:               keyPar5.Name,
				SecretAccessKey:    keyPar5.SecretAccessKey,
			}, pulumi.Provider(providerPa7))
			if err != nil {
				return err
			}

			userKeysPar5[username] = keyPar5
		}

		// ------------------------------------------------------------------
		// Step 10: Remote credentials
		// ------------------------------------------------------------------
		remoteCredsPar5ToPa7, err := flashblade.NewObjectStoreRemoteCredentials(ctx, "par5ToPa7", &flashblade.ObjectStoreRemoteCredentialsArgs{
			Name:              pulumi.String(fmt.Sprintf("%s/%s-creds", pa7ArrayName, bucketName)),
			AccessKeyId:       replicationKeyPa7.AccessKeyId,
			SecretAccessKey:   replicationKeyPa7.SecretAccessKey,
			RemoteName:        pulumi.String(pa7ArrayName),
		}, pulumi.Provider(providerPar5))
		if err != nil {
			return err
		}

		remoteCredsPa7ToPar5, err := flashblade.NewObjectStoreRemoteCredentials(ctx, "pa7ToPar5", &flashblade.ObjectStoreRemoteCredentialsArgs{
			Name:              pulumi.String(fmt.Sprintf("%s/%s-creds", par5ArrayName, bucketName)),
			AccessKeyId:       replicationKeyPar5.AccessKeyId,
			SecretAccessKey:   replicationKeyPar5.SecretAccessKey,
			RemoteName:        pulumi.String(par5ArrayName),
		}, pulumi.Provider(providerPa7))
		if err != nil {
			return err
		}

		// ------------------------------------------------------------------
		// Step 11: Bidirectional replica links
		// ------------------------------------------------------------------
		replicaLinkPar5ToPa7, err := flashblade.NewBucketReplicaLink(ctx, "par5ToPa7", &flashblade.BucketReplicaLinkArgs{
			LocalBucketName:      bucketPar5.Name,
			RemoteBucketName:     bucketPa7.Name,
			RemoteCredentialsName: remoteCredsPar5ToPa7.Name,
		}, pulumi.Provider(providerPar5))
		if err != nil {
			return err
		}

		replicaLinkPa7ToPar5, err := flashblade.NewBucketReplicaLink(ctx, "pa7ToPar5", &flashblade.BucketReplicaLinkArgs{
			LocalBucketName:      bucketPa7.Name,
			RemoteBucketName:     bucketPar5.Name,
			RemoteCredentialsName: remoteCredsPa7ToPar5.Name,
		}, pulumi.Provider(providerPa7))
		if err != nil {
			return err
		}

		// ------------------------------------------------------------------
		// Step 11b: Optional S3 target replication
		// ------------------------------------------------------------------
		if enableS3TargetReplication {
			remoteCredsPar5ToPa7S3, err := flashblade.NewObjectStoreRemoteCredentials(ctx, "par5ToPa7S3", &flashblade.ObjectStoreRemoteCredentialsArgs{
				Name:            pulumi.String(fmt.Sprintf("%s/%s-creds", flashbladeTargetPar5SeesPa7, bucketName)),
				AccessKeyId:     replicationKeyPa7.AccessKeyId,
				SecretAccessKey: replicationKeyPa7.SecretAccessKey,
				TargetName:      pulumi.String(flashbladeTargetPar5SeesPa7),
			}, pulumi.Provider(providerPar5))
			if err != nil {
				return err
			}

			remoteCredsPa7ToPar5S3, err := flashblade.NewObjectStoreRemoteCredentials(ctx, "pa7ToPar5S3", &flashblade.ObjectStoreRemoteCredentialsArgs{
				Name:            pulumi.String(fmt.Sprintf("%s/%s-creds", flashbladeTargetPa7SeesPar5, bucketName)),
				AccessKeyId:     replicationKeyPar5.AccessKeyId,
				SecretAccessKey: replicationKeyPar5.SecretAccessKey,
				TargetName:      pulumi.String(flashbladeTargetPa7SeesPar5),
			}, pulumi.Provider(providerPa7))
			if err != nil {
				return err
			}

			_, err = flashblade.NewBucketReplicaLink(ctx, "par5ToPa7S3", &flashblade.BucketReplicaLinkArgs{
				LocalBucketName:       bucketPar5.Name,
				RemoteBucketName:      bucketPa7.Name,
				RemoteCredentialsName: remoteCredsPar5ToPa7S3.Name,
			}, pulumi.Provider(providerPar5))
			if err != nil {
				return err
			}

			_, err = flashblade.NewBucketReplicaLink(ctx, "pa7ToPar5S3", &flashblade.BucketReplicaLinkArgs{
				LocalBucketName:       bucketPa7.Name,
				RemoteBucketName:      bucketPar5.Name,
				RemoteCredentialsName: remoteCredsPa7ToPar5S3.Name,
			}, pulumi.Provider(providerPa7))
			if err != nil {
				return err
			}
		}

		// ------------------------------------------------------------------
		// Step 12: Lifecycle rules
		// ------------------------------------------------------------------
		for ruleID, ruleCfg := range lifecycleRules {
			safeRule := strings.ReplaceAll(ruleID, "-", "_")

			_, err := flashblade.NewLifecycleRule(ctx, fmt.Sprintf("par5_%s", safeRule), &flashblade.LifecycleRuleArgs{
				BucketName:                        bucketPar5.Name,
				RuleId:                            pulumi.String(ruleID),
				Prefix:                            pulumi.String(ruleCfg.Prefix),
				Enabled:                           pulumi.Bool(ruleCfg.Enabled),
				KeepPreviousVersionFor:            pulumi.IntPtrFromPtr(ruleCfg.KeepPreviousVersionFor),
				KeepCurrentVersionFor:             pulumi.IntPtrFromPtr(ruleCfg.KeepCurrentVersionFor),
				KeepCurrentVersionUntil:           pulumi.IntPtrFromPtr(ruleCfg.KeepCurrentVersionUntil),
				AbortIncompleteMultipartUploadsAfter: pulumi.IntPtrFromPtr(ruleCfg.AbortIncompleteMultipartUploadsAfter),
			}, pulumi.Provider(providerPar5))
			if err != nil {
				return err
			}

			_, err = flashblade.NewLifecycleRule(ctx, fmt.Sprintf("pa7_%s", safeRule), &flashblade.LifecycleRuleArgs{
				BucketName:                        bucketPa7.Name,
				RuleId:                            pulumi.String(ruleID),
				Prefix:                            pulumi.String(ruleCfg.Prefix),
				Enabled:                           pulumi.Bool(ruleCfg.Enabled),
				KeepPreviousVersionFor:            pulumi.IntPtrFromPtr(ruleCfg.KeepPreviousVersionFor),
				KeepCurrentVersionFor:             pulumi.IntPtrFromPtr(ruleCfg.KeepCurrentVersionFor),
				KeepCurrentVersionUntil:           pulumi.IntPtrFromPtr(ruleCfg.KeepCurrentVersionUntil),
				AbortIncompleteMultipartUploadsAfter: pulumi.IntPtrFromPtr(ruleCfg.AbortIncompleteMultipartUploadsAfter),
			}, pulumi.Provider(providerPa7))
			if err != nil {
				return err
			}
		}

		// ------------------------------------------------------------------
		// Step 13: Audit filters (optional)
		// ------------------------------------------------------------------
		if auditEnabled {
			_, err := flashblade.NewBucketAuditFilter(ctx, "par5", &flashblade.BucketAuditFilterArgs{
				Name:        pulumi.String("auditwrite"),
				BucketName:  bucketPar5.Name,
				Actions:     pulumi.ToStringArray([]string{"s3:PutObject", "s3:DeleteObject"}),
				S3Prefixes:  pulumi.ToStringArray(auditPrefixes),
			}, pulumi.Provider(providerPar5))
			if err != nil {
				return err
			}

			_, err = flashblade.NewBucketAuditFilter(ctx, "pa7", &flashblade.BucketAuditFilterArgs{
				Name:        pulumi.String("auditwrite"),
				BucketName:  bucketPa7.Name,
				Actions:     pulumi.ToStringArray([]string{"s3:PutObject", "s3:DeleteObject"}),
				S3Prefixes:  pulumi.ToStringArray(auditPrefixes),
			}, pulumi.Provider(providerPa7))
			if err != nil {
				return err
			}
		}

		// ------------------------------------------------------------------
		// Step 14: QoS policy (optional)
		// ------------------------------------------------------------------
		if qos != nil {
			_, err := flashblade.NewQosPolicy(ctx, "this", &flashblade.QosPolicyArgs{
				Name:                  pulumi.String(fmt.Sprintf("%s-qos", accountName)),
				Enabled:               pulumi.Bool(true),
				MaxTotalBytesPerSec:   pulumi.IntPtrFromPtr(qos.MaxTotalBytesPerSec),
				MaxTotalOpsPerSec:     pulumi.IntPtrFromPtr(qos.MaxTotalOpsPerSec),
			}, pulumi.Provider(providerPar5))
			if err != nil {
				return err
			}
		}

		// ------------------------------------------------------------------
		// Outputs
		// ------------------------------------------------------------------
		ctx.Export("par5ConnectionStatus", pulumi.String(par5SeesPa7.Status))
		ctx.Export("pa7ConnectionStatus", pulumi.String(pa7SeesPar5.Status))
		ctx.Export("par5BucketId", bucketPar5.ID())
		ctx.Export("pa7BucketId", bucketPa7.ID())
		ctx.Export("par5ReplicaStatus", replicaLinkPar5ToPa7.Status)
		ctx.Export("pa7ReplicaStatus", replicaLinkPa7ToPar5.Status)
		ctx.Export("fqdn", pulumi.String(fqdn))

		// Per-user access key IDs
		userKeyIDs := make(map[string]pulumi.StringOutput)
		for username, key := range userKeysPar5 {
			userKeyIDs[username] = key.AccessKeyId
		}
		ctx.Export("userAccessKeyIds", pulumi.ToStringMapOutput(userKeyIDs))

		return nil
	})
}

func intPtr(i int) *int {
	return &i
}
