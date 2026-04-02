import logging
import subprocess
import shutil
import time
from pathlib import Path
from typing import Tuple

TEST_DATA = (Path(__file__).parent.parent / "test-data").resolve()
COMPOSE_FILE = (Path(__file__).parent.parent / "test-compose.yaml").resolve()
KEYS_DIR = TEST_DATA / "keys"
REPOS_DIR = TEST_DATA / "repos"
LOGS_DIR = TEST_DATA / "logs"


class TestE2E:
  @classmethod
  def exec_in_host(cls, cmd: list[str]) -> None:
    logging.info(f"exec_in_host: {' '.join(cmd)}")
    subprocess.run(cmd, check=True)

  @classmethod
  def generate_ssh_key(cls, key_file: Path):
    cls.exec_in_host(
      [
        "ssh-keygen",
        "-t",
        "ed25519",
        "-f",
        str(key_file),
        "-N",
        "",
        "-C",
        "e2e-test",
        "-q",
      ],
    )

  @classmethod
  def stop_docker_compose(cls):
    cls.exec_in_host(
      [
        "docker",
        "compose",
        "-f",
        str(COMPOSE_FILE),
        "down",
        "-v",
        "--remove-orphans",
      ],
    )

  @classmethod
  def setup_class(cls):
    assert TEST_DATA.exists()

    # clean up from previous runs
    cls.stop_docker_compose()
    shutil.rmtree(KEYS_DIR, ignore_errors=True)
    shutil.rmtree(REPOS_DIR, ignore_errors=True)
    shutil.rmtree(LOGS_DIR, ignore_errors=True)

    KEYS_DIR.mkdir(exist_ok=False)
    REPOS_DIR.mkdir(exist_ok=False)
    LOGS_DIR.mkdir(exist_ok=False)

    key_file = KEYS_DIR / "id_ed25519"

    cls.generate_ssh_key(key_file)

    repo_dir = REPOS_DIR / "repo1"
    repo_dir.mkdir(parents=True)
    cls.exec_in_host(["git", "init", "--bare", str(repo_dir)])
    cls.exec_in_host(["docker", "compose", "-f", str(COMPOSE_FILE), "up", "-d", "--build"])
    time.sleep(3)

  @classmethod
  def teardown_class(cls):
    cls.stop_docker_compose()

  @staticmethod
  def exec_in_test_app(cmd: str) -> Tuple[int, str, str]:
    logging.info(f"exec_in_test_app: {cmd}")
    result = subprocess.run(
      [
        "docker",
        "compose",
        "-f",
        str(COMPOSE_FILE),
        "exec",
        "-T",
        "--user=1000",
        "test-app",
        "/bin/sh",
        "-c",
        cmd,
      ],
      capture_output=True,
      text=True,
      check=True,
    )
    return result.returncode, result.stdout, result.stderr

  def test_git_clone_from_allowed_host(self):
    self.exec_in_test_app("rm -rf /tmp/repo1")
    code, stdout, stderr = self.exec_in_test_app(
      "git clone git@git-server:/git-server/repos/repo1 /tmp/repo1"
    )
    assert code == 0, f"git clone failed: {stderr}"

  def test_git_commit_and_push(self):
    code, stdout, stderr = self.exec_in_test_app(
      "cd /tmp/repo1 && echo 'e2e marker' > marker.txt && "
      "git add marker.txt && "
      "git commit -m 'e2e: marker commit' && "
      "git push origin HEAD"
    )
    assert code == 0, f"git push failed: {stderr}"

  def test_git_pull(self):
    code, stdout, stderr = self.exec_in_test_app("cd /tmp/repo1 && git pull")
    assert code == 0, f"git pull failed: {stderr}"

  def test_ssh_to_disallowed_host_blocked(self):
    code, stdout, stderr = self.exec_in_test_app("ssh git@not-allowed-host echo hi 2>&1 || true")
    combined = stdout + stderr
    assert "Access Denied" in combined, f"Disallowed host was not blocked: {combined}"

  def test_log_file_exists(self):
    log_file = LOGS_DIR / "ssh-wrapper.log"
    assert log_file.exists(), f"Log file missing: {log_file}"
    assert log_file.stat().st_size > 0, f"Log file is empty: {log_file}"

  def test_log_contains_git_server_activity(self):
    log_file = LOGS_DIR / "ssh-wrapper.log"
    content = log_file.read_text()
    assert "git-server" in content, f"Log does not contain git-server activity: {content}"
