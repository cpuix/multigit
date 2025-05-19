class Multigit < Formula
  desc "A tool for managing multiple git repositories with ease"
  homepage "https://github.com/cpuix/multigit"
  version "0.1.0"
  license "MIT"

  if OS.mac? && Hardware::CPU.intel?
    url "https://github.com/cpuix/multigit/releases/download/v#{version}/multigit-darwin-amd64.tar.gz"
    sha256 "" # Replace with actual SHA-256 checksum
  elsif OS.mac? && Hardware::CPU.arm?
    url "https://github.com/cpuix/multigit/releases/download/v#{version}/multigit-darwin-arm64.tar.gz"
    sha256 "" # Replace with actual SHA-256 checksum
  elsif OS.linux? && Hardware::CPU.intel?
    url "https://github.com/cpuix/multigit/releases/download/v#{version}/multigit-linux-amd64.tar.gz"
    sha256 "" # Replace with actual SHA-256 checksum
  elsif OS.linux? && Hardware::CPU.arm? && Hardware::CPU.is_64_bit?
    url "https://github.com/cpuix/multigit/releases/download/v#{version}/multigit-linux-arm64.tar.gz"
    sha256 "" # Replace with actual SHA-256 checksum
  end

  def install
    bin.install "multigit"
  end

  test do
    system "#{bin}/multigit", "--version"
  end
end
