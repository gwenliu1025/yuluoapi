package repository

import (
	"context"
	"testing"
	"testing/fstest"

	"github.com/DATA-DOG/go-sqlmock"
	dbmigrations "github.com/Wei-Shaw/sub2api/migrations"
	"github.com/stretchr/testify/require"
)

const unknownMigrationChecksum = "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"

func TestIsMigrationChecksumCompatible(t *testing.T) {
	t.Run("054历史checksum可兼容", func(t *testing.T) {
		ok := isMigrationChecksumCompatible(
			"054_drop_legacy_cache_columns.sql",
			"182c193f3359946cf094090cd9e57d5c3fd9abaffbc1e8fc378646b8a6fa12b4",
			"82de761156e03876653e7a6a4eee883cd927847036f779b0b9f34c42a8af7a7d",
		)
		require.True(t, ok)
	})

	t.Run("054在未知文件checksum下不兼容", func(t *testing.T) {
		ok := isMigrationChecksumCompatible(
			"054_drop_legacy_cache_columns.sql",
			"182c193f3359946cf094090cd9e57d5c3fd9abaffbc1e8fc378646b8a6fa12b4",
			"0000000000000000000000000000000000000000000000000000000000000000",
		)
		require.False(t, ok)
	})

	t.Run("061历史checksum可兼容", func(t *testing.T) {
		ok := isMigrationChecksumCompatible(
			"061_add_usage_log_request_type.sql",
			"08a248652cbab7cfde147fc6ef8cda464f2477674e20b718312faa252e0481c0",
			"66207e7aa5dd0429c2e2c0fabdaf79783ff157fa0af2e81adff2ee03790ec65c",
		)
		require.True(t, ok)
	})

	t.Run("061第二个历史checksum可兼容", func(t *testing.T) {
		ok := isMigrationChecksumCompatible(
			"061_add_usage_log_request_type.sql",
			"222b4a09c797c22e5922b6b172327c824f5463aaa8760e4f621bc5c22e2be0f3",
			"66207e7aa5dd0429c2e2c0fabdaf79783ff157fa0af2e81adff2ee03790ec65c",
		)
		require.True(t, ok)
	})

	t.Run("非白名单迁移不兼容", func(t *testing.T) {
		ok := isMigrationChecksumCompatible(
			"001_init.sql",
			"182c193f3359946cf094090cd9e57d5c3fd9abaffbc1e8fc378646b8a6fa12b4",
			"82de761156e03876653e7a6a4eee883cd927847036f779b0b9f34c42a8af7a7d",
		)
		require.False(t, ok)
	})

	t.Run("109历史checksum可兼容", func(t *testing.T) {
		ok := isMigrationChecksumCompatible(
			"109_auth_identity_compat_backfill.sql",
			"551e498aa5616d2d91096e9d72cf9fb36e418ee22eacc557f8811cadbc9e20ee",
			"0580b4602d85435edf9aca1633db580bb3932f26517f75134106f80275ec2ace",
		)
		require.True(t, ok)
	})

	t.Run("109当前checksum可兼容历史checksum", func(t *testing.T) {
		ok := isMigrationChecksumCompatible(
			"109_auth_identity_compat_backfill.sql",
			"551e498aa5616d2d91096e9d72cf9fb36e418ee22eacc557f8811cadbc9e20ee",
			"0580b4602d85435edf9aca1633db580bb3932f26517f75134106f80275ec2ace",
		)
		require.True(t, ok)
	})

	t.Run("109回滚到历史文件后仍兼容已应用的新checksum", func(t *testing.T) {
		ok := isMigrationChecksumCompatible(
			"109_auth_identity_compat_backfill.sql",
			"0580b4602d85435edf9aca1633db580bb3932f26517f75134106f80275ec2ace",
			"551e498aa5616d2d91096e9d72cf9fb36e418ee22eacc557f8811cadbc9e20ee",
		)
		require.True(t, ok)
	})

	t.Run("110历史checksum可兼容", func(t *testing.T) {
		ok := isMigrationChecksumCompatible(
			"110_pending_auth_and_provider_default_grants.sql",
			"e3d1f433be2b564cfbdc549adf98fce13c5c7b363ebc20fd05b765d0563b0925",
			"32cf87ee787b1bb36b5c691367c96eee37518fa3eed6f3322cf68795e3745279",
		)
		require.True(t, ok)
	})

	t.Run("112历史checksum可兼容", func(t *testing.T) {
		ok := isMigrationChecksumCompatible(
			"112_add_payment_order_provider_key_snapshot.sql",
			"ffd3e8a2c9295fa9cbefefd629a78268877e5b51bc970a82d9b3f46ec4ebd15e",
			"b75f8f56d39455682787696a3d92ad25b055444ca328fb7fca9a460a15d68d99",
		)
		require.True(t, ok)
	})

	t.Run("115历史checksum可兼容修复后的legacy external backfill", func(t *testing.T) {
		ok := isMigrationChecksumCompatible(
			"115_auth_identity_legacy_external_backfill.sql",
			"4cf39e508be9fd1a5aa41610cbbebeb80385c9adda45bf78a706de9db4f1385f",
			"022aadd97bb53e755f0cf7a3a957e0cb1a1353b0c39ec4de3234acd2871fd04f",
		)
		require.True(t, ok)
	})

	t.Run("116历史checksum可兼容修复后的legacy external safety reports", func(t *testing.T) {
		ok := isMigrationChecksumCompatible(
			"116_auth_identity_legacy_external_safety_reports.sql",
			"f7757bd929ac67ffb08ce69fa4cf20fad39dbff9d5a5085fb2adabb7607e5877",
			"07edb09fa8d04ffb172b0621e3c22f4d1757d20a24ae267b3b36b087ab72d488",
		)
		require.True(t, ok)
	})

	t.Run("119历史checksum可兼容占位文件", func(t *testing.T) {
		ok := isMigrationChecksumCompatible(
			"119_enforce_payment_orders_out_trade_no_unique.sql",
			"ebd2c67cce0116393fb4f1b5d5116a67c6aceb73820dfb5133d1ff6f36d72d34",
			"0bbe809ae48a9d811dabda1ba1c74955bd71c4a9cc610f9128816818dfa6c11e",
		)
		require.True(t, ok)
	})

	t.Run("118多个历史checksum都可兼容当前版本", func(t *testing.T) {
		for _, dbChecksum := range []string{
			"a38243ca0a72c3a01c0a92b7986423054d6133c0399441f853b99802852720fb",
			"e0cdf835d6c688d64100f483d31bc02ac9ebad414bf1837af239a84bf75b8227",
		} {
			ok := isMigrationChecksumCompatible(
				"118_wechat_dual_mode_and_auth_source_defaults.sql",
				dbChecksum,
				"b54194d7a3e4fbf710e0a3590d22a2fe7966804c487052a356e0b55f53ef96b0",
			)
			require.True(t, ok)
		}
	})

	t.Run("120多个历史checksum都可兼容新的notx修复版本", func(t *testing.T) {
		for _, dbChecksum := range []string{
			"e77921f79d539bc24575cb9c16cbe566d2b23ce816190343d0a7568f6a3fcf61",
			"707431450603e70a43ce9fbd61e0c12fa67da4875158ccefabacea069587ab22",
			"04b082b5a239c525154fe9185d324ee2b05ff90da9297e10dba19f9be79aa59a",
		} {
			ok := isMigrationChecksumCompatible(
				"120_enforce_payment_orders_out_trade_no_unique_notx.sql",
				dbChecksum,
				"34aadc0db59a4e390f92a12b73bd74642d9724f33124f73638ae00089ea5e074",
			)
			require.True(t, ok)
		}
	})

	t.Run("119未知checksum不兼容", func(t *testing.T) {
		ok := isMigrationChecksumCompatible(
			"119_enforce_payment_orders_out_trade_no_unique.sql",
			"ebd2c67cce0116393fb4f1b5d5116a67c6aceb73820dfb5133d1ff6f36d72d34",
			"ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
		)
		require.False(t, ok)
	})
}

func TestPublishedMigrationChecksumUpgradeCompatibility(t *testing.T) {
	tests := []struct {
		name            string
		v020Checksum    string
		v02001Checksum  string
		legacyChecksums []string
	}{
		{
			name:           "142_user_platform_quotas.sql",
			v020Checksum:   "fe485a0663f8948174819a2d36b06f260c2dd8a8e85b6e39cd85798b75090873",
			v02001Checksum: "ba6af6ac4442770132844bb301260019d87497bad216afe420db9247dfa78d58",
		},
		{
			name:           "159_batch_image_foundation.sql",
			v020Checksum:   "d902b70982025ec519749faf058aab7631e82c3f48167b9a4ae4db718eb72cce",
			v02001Checksum: "e8444330ab588374c3df3fca7d7029c249166504aa74789f5c5d55159cb6f96a",
			legacyChecksums: []string{
				"82da85b5d98e67a0507647b873a40373e84538e4adafdeed6767c0ac8b6570b2",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			content, err := dbmigrations.FS.ReadFile(tc.name)
			require.NoError(t, err)

			require.NoError(t, runPublishedMigrationChecksumCheck(t, tc.name, content, tc.v020Checksum))
			require.Equal(t, tc.v020Checksum, migrationChecksum(string(content)))
			require.NoError(t, runPublishedMigrationChecksumCheck(t, tc.name, content, tc.v02001Checksum))
			for _, checksum := range tc.legacyChecksums {
				require.NoError(t, runPublishedMigrationChecksumCheck(t, tc.name, content, checksum))
			}

			err = runPublishedMigrationChecksumCheck(t, tc.name, content, unknownMigrationChecksum)
			require.ErrorContains(t, err, "checksum mismatch")
		})
	}
}

func TestBatchImageCurrencyChangeUsesFollowupMigration(t *testing.T) {
	foundation, err := dbmigrations.FS.ReadFile("159_batch_image_foundation.sql")
	require.NoError(t, err)
	require.Contains(t, string(foundation), "currency VARCHAR(16) NOT NULL DEFAULT 'USD'")

	followup, err := dbmigrations.FS.ReadFile("234_billing_currency_cny.sql")
	require.NoError(t, err)
	require.Contains(t, string(followup), "ALTER COLUMN currency SET DEFAULT 'CNY'")
}

func runPublishedMigrationChecksumCheck(t *testing.T, name string, content []byte, dbChecksum string) error {
	t.Helper()
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	prepareMigrationsBootstrapExpectations(mock)
	mock.ExpectQuery("SELECT checksum FROM schema_migrations WHERE filename = \\$1").
		WithArgs(name).
		WillReturnRows(sqlmock.NewRows([]string{"checksum"}).AddRow(dbChecksum))
	mock.ExpectExec("SELECT pg_advisory_unlock\\(\\$1\\)").
		WithArgs(migrationsAdvisoryLockID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = applyMigrationsFS(context.Background(), db, fstest.MapFS{
		name: &fstest.MapFile{Data: content},
	})
	require.NoError(t, mock.ExpectationsWereMet())
	return err
}
