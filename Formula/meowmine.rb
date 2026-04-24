class Meowmine < Formula
  desc "Cat-themed crypto mining tycoon TUI game (Bubbletea + Lipgloss)"
  homepage "https://github.com/RandomNameORG/kitten-crypto-mining-ventures"
  version "0.1.0"
  license "MIT"

  base = "https://github.com/RandomNameORG/kitten-crypto-mining-ventures/releases/download/v#{version}"

  on_macos do
    on_arm do
      url "#{base}/meowmine-darwin-arm64"
      sha256 "6a2420193561084015f27201d90c6518e1724f32be995a48db53d6ee117b6b8e"

      resource "meowmine-ssh" do
        url "#{base}/meowmine-ssh-darwin-arm64"
        sha256 "8d6f8d9475f4d61fb5595353a718a14553c47d0ff45a1c752768ebf911fb48f2"
      end
    end

    on_intel do
      url "#{base}/meowmine-darwin-amd64"
      sha256 "ced6725057e2c63065ec1f374cb6f6cc782bf1f890ec6d2fc0debf30c9e4248e"

      resource "meowmine-ssh" do
        url "#{base}/meowmine-ssh-darwin-amd64"
        sha256 "6d18845307666c4df0b534c23a70c6b45480b9547ed36b49fbd7d28f7c90908a"
      end
    end
  end

  on_linux do
    on_intel do
      url "#{base}/meowmine-linux-amd64"
      sha256 "dbe6429054c3528401e8f865b3a94123d9536114dea37d8e50f6bbd80a5cd2da"

      resource "meowmine-ssh" do
        url "#{base}/meowmine-ssh-linux-amd64"
        sha256 "cdeef2c8b83426ed7a05a89bd3ea9a33e48c923bb5ef03bc4f5691d83971d7fa"
      end
    end
  end

  def install
    local_name = Dir["meowmine-*"].find { |f| !f.start_with?("meowmine-ssh") }
    bin.install local_name => "meow"

    resource("meowmine-ssh").stage do
      ssh_name = Dir["meowmine-ssh-*"].first
      bin.install ssh_name => "meow-ssh"
    end
  end

  test do
    assert_match "-new", shell_output("#{bin}/meow -h 2>&1")
    assert_match "-port", shell_output("#{bin}/meow-ssh -h 2>&1")
  end
end
