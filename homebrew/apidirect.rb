class Apidirect < Formula
  desc "CLI tool for rapid API deployment and marketplace management"
  homepage "https://github.com/api-direct/cli"
  version "1.0.0"
  
  on_macos do
    if Hardware::CPU.intel?
      url "https://github.com/api-direct/cli/releases/download/v#{version}/apidirect_#{version}_darwin_x86_64.tar.gz"
      sha256 "PLACEHOLDER_SHA256_DARWIN_AMD64"
    elsif Hardware::CPU.arm?
      url "https://github.com/api-direct/cli/releases/download/v#{version}/apidirect_#{version}_darwin_arm64.tar.gz"
      sha256 "PLACEHOLDER_SHA256_DARWIN_ARM64"
    end
  end
  
  on_linux do
    if Hardware::CPU.intel?
      if Hardware::CPU.is_64_bit?
        url "https://github.com/api-direct/cli/releases/download/v#{version}/apidirect_#{version}_linux_x86_64.tar.gz"
        sha256 "PLACEHOLDER_SHA256_LINUX_AMD64"
      else
        url "https://github.com/api-direct/cli/releases/download/v#{version}/apidirect_#{version}_linux_i386.tar.gz"
        sha256 "PLACEHOLDER_SHA256_LINUX_386"
      end
    elsif Hardware::CPU.arm?
      url "https://github.com/api-direct/cli/releases/download/v#{version}/apidirect_#{version}_linux_arm64.tar.gz"
      sha256 "PLACEHOLDER_SHA256_LINUX_ARM64"
    end
  end
  
  def install
    bin.install "apidirect"
    
    # Install shell completions
    generate_completions_from_executable(bin/"apidirect", "completion")
  end
  
  test do
    assert_match "API Direct CLI", shell_output("#{bin}/apidirect --version")
    assert_match "Import existing API projects", shell_output("#{bin}/apidirect --help")
  end
end