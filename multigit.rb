#!/usr/bin/env ruby

# Multigt is a tool to manage multiple GitHub accounts on the same machine.
# It allows you to create, delete, copy and use SSH keys for different GitHub accounts.
# Creator: @cpuix

require 'fileutils'
require_relative 'validation'
require_relative 'localization'

class MultiGit
  SSH_CONFIG_PATH = File.join(ENV['HOME'], '.ssh', 'config')

  @debug_mode = false

  def self.debug_mode=(mode)
    @debug_mode = mode
  end

  def self.debug_mode
    @debug_mode
  end

  def initialize(command, *args)
    @command = command
    @args = args
    @localization = Localization.new
  end

  def execute
    begin
      case @command
      when 'create'
        create_account_key(*@args)
      when 'delete'
        delete_account_key(*@args)
      when 'copy'
        copy_public_key(*@args)
      when 'use'
        use_account(*@args)
      else
        raise ArgumentError, @localization.get_message("system.incorrect_command")
      end
    rescue Interrupt
      puts @localization.get_message("system.operation_cancelled")
      exit 0
    rescue StandardError => e
      if self.class.debug_mode
        puts e.full_message(highlight: true, order: :top)
      else
        puts e.message
      end
      exit 1
    end
  end

  private

  def create_account_key(*args)
    account_name, user_email = args

    if account_name.nil? || account_name.empty?
      puts @localization.get_message("input.account_name")
      account_name = STDIN.gets.chomp
    end

    if user_email.nil? || user_email.empty?
      puts @localization.get_message("input.email")
      user_email = STDIN.gets.chomp
    end

    unless Validation.valid_account_name?(account_name)
      puts @localization.get_message("error.invalid_account_name")
      exit 1
    end

    unless Validation.valid_email?(user_email)
      puts @localization.get_message("error.invalid_email")
      exit 1
    end

    ssh_key_file = File.join(ENV['HOME'], '.ssh', "github-#{account_name}")

    if File.exist?(ssh_key_file)
      puts  @localization.get_message("ssh.key_exists")
      exit 1
    end

    puts @localization.get_message("ssh.add_passphrase")
    add_passphrase = STDIN.gets.chomp

    if add_passphrase == 'y'
      system("ssh-keygen -t ed25519 -C '#{user_email}' -f #{ssh_key_file}")
    else
      system("ssh-keygen -t ed25519 -C '#{user_email}' -f #{ssh_key_file} -N ''")
    end

    system("ssh-add -K #{ssh_key_file}")

    config_entry = <<~CONFIG
      Host github.com-#{account_name}
      HostName github.com
      User git
      IdentityFile #{ssh_key_file}
    CONFIG

    add_ssh_config_entry(config_entry, account_name, ssh_key_file)

    puts @localization.get_message("ssh.created")
    puts @localization.get_message("ssh.copy_public_key")
    copy_file = File.read("#{ssh_key_file}.pub")
    IO.popen('pbcopy', 'w') { |f| f << copy_file }
    puts copy_file
  end

  def delete_account_key(*args)
    account_name = args

    if account_name.nil? || account_name.empty?
      puts @localization.get_message("input.account_name")
      account_name = STDIN.gets.chomp
    end

    unless Validation.valid_account_name?(account_name)
      puts @localization.get_message("error.invalid_account_name")
      exit 1
    end

    ssh_key_file = File.join(ENV['HOME'], '.ssh', "github-#{account_name}")

    unless File.exist?(ssh_key_file)
      puts @localization.get_message("error.key_file_not_found")
      exit 1
    end

    puts @localization.get_message("input.confirm_delete")
    confirm_delete = STDIN.gets.chomp

    if confirm_delete == 'y'
      FileUtils.rm_rf([ssh_key_file, "#{ssh_key_file}.pub"])
      File.write(SSH_CONFIG_PATH, File.read(SSH_CONFIG_PATH).gsub(/Host github.com-#{account_name}\nHostName github.com\nUser git\nIdentityFile #{ssh_key_file}\n/, ''))
      puts @localization.get_message("ssh.deleted")
    else
      puts @localization.get_message("system.operation_cancelled")
    end
  end

  def copy_public_key(*args)
    account_name = args

    if account_name.nil? || account_name.empty?
      puts @localization.get_message("input.account_name")
      account_name = STDIN.gets.chomp
    end

    unless Validation.valid_account_name?(account_name)
      puts @localization.get_message("error.invalid_account_name")
      exit 1
    end

    ssh_key_file = File.join(ENV['HOME'], '.ssh', "github-#{account_name}")

    unless File.exist?(ssh_key_file)
      puts @localization.get_message("error.key_file_not_found")
      exit 1
    end

    IO.popen('pbcopy', 'w') { |f| f << File.read("#{ssh_key_file}.pub") }
    puts "github-#{account_name} Public Key copied."
  end

  def use_account(name)
    # use_account metodunun içeriği...
  end

  def add_ssh_config_entry(config_entry, account_name, ssh_key_file)
    existing_entry = File.read(SSH_CONFIG_PATH).match(/Host github.com-#{account_name}\nHostName github.com\nUser git\nIdentityFile #{ssh_key_file}/)

    if existing_entry.nil?
      File.open(SSH_CONFIG_PATH, 'a') { |f| f.write("\n#{config_entry}") }
    else
      File.write(SSH_CONFIG_PATH, File.read(SSH_CONFIG_PATH).gsub(existing_entry.to_s, config_entry))
    end
  end
end

if __FILE__ == $0
  if ARGV.empty?
    puts "Usage: multigit <command> [<args>]"
    puts "Commands:"
    puts "  create <account_name> <email> - Create a new SSH key for a GitHub account."
    puts "  delete <account_name> - Delete a SSH key for a GitHub account."
    puts "  copy <account_name> - Copy the public key for a GitHub account."
    puts "  use <account_name> - Use a GitHub account."

    exit 1
  end

  command = ARGV.shift
  multigit = MultiGit.new(command, *ARGV)
  multigit.execute
end