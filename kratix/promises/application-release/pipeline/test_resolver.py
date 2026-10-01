import os
import subprocess
import tempfile
import unittest
from pathlib import Path

import yaml

import resolver

REPO_ROOT = Path(__file__).resolve().parents[4]
PROFILE = REPO_ROOT / "artifacts" / "request-logger-http.profile.yaml"
BASE_MANIFEST = REPO_ROOT / "kratix" / "manifests" / "apps" / "request-logger-application-release.base.yaml"
MATERIALIZE = REPO_ROOT / "kratix" / "scripts" / "materialize-application-release.sh"
CATALOG = REPO_ROOT / "kratix" / "manifests" / "catalog" / "todos-api-catalog.yaml"
BREAKING_CATALOG = REPO_ROOT / "kratix" / "manifests" / "catalog" / "todos-api-catalog-breaking.yaml"


class ResolverTestCase(unittest.TestCase):
    def setUp(self):
        self.tmp = tempfile.TemporaryDirectory()
        self.addCleanup(self.tmp.cleanup)
        self.use_catalog(CATALOG)

    def use_catalog(self, configmap_path):
        catalog_dir = Path(self.tmp.name) / f"catalog-{configmap_path.stem}"
        catalog_dir.mkdir(exist_ok=True)
        for key, value in yaml.safe_load(configmap_path.read_text())["data"].items():
            (catalog_dir / key).write_text(value)
        os.environ["RUNTIME_CONDITIONS_CATALOG_DIR"] = str(catalog_dir)
        self.addCleanup(os.environ.pop, "RUNTIME_CONDITIONS_CATALOG_DIR", None)

    def materialized_release(self):
        output = Path(self.tmp.name) / "release.yaml"
        subprocess.run(
            [str(MATERIALIZE), str(BASE_MANIFEST), str(PROFILE), str(output)],
            check=True,
            capture_output=True,
        )
        return yaml.safe_load(output.read_text())

    def release_with_conditions(self, conditions):
        release = self.materialized_release()
        profile = yaml.safe_load(release["spec"]["profile"])
        profile["conditions"] = conditions
        release["spec"]["profile"] = yaml.safe_dump(profile)
        return release

    def canonical_conditions(self):
        return yaml.safe_load(PROFILE.read_text())["conditions"]


class MaterializationTest(ResolverTestCase):
    def test_generated_profile_is_embedded_unchanged(self):
        release = self.materialized_release()
        self.assertEqual(release["spec"]["profile"], PROFILE.read_text())

    def test_materialized_release_resolves_to_platform_resources(self):
        output = resolver.resolve(self.materialized_release())
        by_kind = {doc["kind"]: doc for doc in output}

        self.assertEqual(by_kind["Redis"]["metadata"]["name"], "request-logger-cache")
        env = {
            var["name"]: var["value"]
            for var in by_kind["Deployment"]["spec"]["template"]["spec"]["containers"][0]["env"]
        }
        self.assertEqual(
            env,
            {
                "TODOS_API_URL": "http://todos-api.demo.svc.cluster.local:8080",
                "REDIS_HOST": "request-logger-cache.demo.svc.cluster.local",
                "REDIS_PORT": "6379",
            },
        )


class SupportEvaluationTest(ResolverTestCase):
    def test_rejects_unknown_condition_kind(self):
        conditions = self.canonical_conditions() + [
            {
                "name": "site-analytics",
                "kind": "google.analytics",
                "interface": {"type": "web"},
                "configuration": {"env": [{"property": "measurementId", "name": "GA_MEASUREMENT_ID"}]},
            }
        ]
        with self.assertRaises(resolver.UnsupportedConditionError) as ctx:
            resolver.resolve(self.release_with_conditions(conditions))
        self.assertEqual([c.name for c in ctx.exception.conditions], ["site-analytics"])
        self.assertIn("'google.analytics'", str(ctx.exception))

    def test_rejects_non_redis_cache_instead_of_skipping_it(self):
        conditions = self.canonical_conditions()
        conditions[1]["interface"]["engine"] = "memcached"
        with self.assertRaises(resolver.UnsupportedConditionError) as ctx:
            resolver.resolve(self.release_with_conditions(conditions))
        self.assertEqual([c.name for c in ctx.exception.conditions], ["request-cache"])

    def test_rejects_property_the_platform_cannot_provide(self):
        conditions = self.canonical_conditions()
        conditions[0]["configuration"]["env"][0]["property"] = "token"
        with self.assertRaises(resolver.UnsupportedConditionError) as ctx:
            resolver.resolve(self.release_with_conditions(conditions))
        self.assertIn("'token'", str(ctx.exception))

    def test_reports_every_unsupported_condition(self):
        conditions = self.canonical_conditions()
        conditions[0]["kind"] = "queue"
        conditions[1]["interface"]["engine"] = "memcached"
        with self.assertRaises(resolver.UnsupportedConditionError) as ctx:
            resolver.resolve(self.release_with_conditions(conditions))
        self.assertEqual(len(ctx.exception.conditions), 2)

    def test_falls_back_to_first_satisfiable_alternative(self):
        conditions = self.canonical_conditions()
        conditions[1]["configuration"]["alternatives"] = [
            {"env": [{"property": "url", "name": "REDIS_URL"}]},
        ]
        output = resolver.resolve(self.release_with_conditions(conditions))
        deployment = next(doc for doc in output if doc["kind"] == "Deployment")
        names = [var["name"] for var in deployment["spec"]["template"]["spec"]["containers"][0]["env"]]
        self.assertEqual(names, ["TODOS_API_URL", "REDIS_URL"])


class ErrorCategoryTest(ResolverTestCase):
    def test_malformed_profile_is_a_profile_error(self):
        release = self.materialized_release()
        release["spec"]["profile"] = "apiVersion: v1\nkind: ConfigMap\nconditions:\n  - kind: api\n"
        with self.assertRaises(resolver.ProfileError):
            resolver.resolve(release)

    def test_breaking_catalog_is_still_a_contract_error(self):
        self.use_catalog(BREAKING_CATALOG)
        with self.assertRaises(resolver.ContractError):
            resolver.resolve(self.materialized_release())


if __name__ == "__main__":
    unittest.main()
