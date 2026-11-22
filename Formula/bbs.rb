class Bbs < Formula
  desc "Convert legacy BBS color-coded textfiles to ANSI, HTML, PNG, or plain text"
  homepage "https://github.com/bengarrett/bbs"
  url "https://github.com/bengarrett/bbs/archive/refs/tags/v1.1.0.tar.gz"
  sha256 "f9a5569adfc1da232d4ec4c74d58d654d5c8deebb21f8d56aca09bf8259c3370"
  version "1.1.0"
  license "MIT"

  livecheck do
    url :stable
    strategy :github_latest
  end

  depends_on "go" => :build

  def install
    ldflags = "-s -w -X main.version=#{version} -X main.date=#{time.iso8601}"
    system "go", "build", *std_go_args(ldflags: ldflags), "./cmd/bbs"
  end

  test do
    assert_match "bbs", shell_output("#{bin}/bbs --version")
  end
end
