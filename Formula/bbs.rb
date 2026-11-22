class Bbs < Formula
  desc "Convert legacy BBS color-coded textfiles to ANSI, HTML, PNG, or plain text"
  homepage "https://github.com/bengarrett/bbs"
  url "https://github.com/bengarrett/bbs/archive/refs/tags/v1.1.1.tar.gz"
  sha256 "e8d588f978ca929dc4559997cfcee32399ea0a0aff031ba39e146461203bc9e7"
  version "1.1.1"
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
