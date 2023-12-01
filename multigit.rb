#!/usr/bin/env ruby

# Multigit is a tool to manage multiple GitHub accounts on the same machine.
# It allows you to create, delete, copy and use SSH keys for different GitHub accounts.

require 'fileutils'
require 'open3'
require 'optparse'
require 'colorize'
require_relative 'modules/localization'
require_relative 'modules/key_manager'
require_relative 'modules/validation'
require_relative 'modules/input_manager'

class MultiGit
  VERSION = '1.0.0'
  class << self
    attr_accessor :debug_mode
  end

  attr_reader :config

  def initialize(command, *args)
    @config = load_config # For example: @config[:ssh_config_path]
    @command = command
    @args = args
    @key_manager = KeyManager.new(load_config, Localization)
  end

  def self.parse_args(args)
    options = { command: nil, args: [] }

    opt_parser = OptionParser.new do |opts|
      opts.banner = "Usage: multigit <command> [options] [args]"
      opts.on("--help", "Prints this help") do
        puts opts
        exit
      end
      opts.on("--version", "Prints version information") do
        puts "multigit v#{VERSION}"
        exit
      end
      opts.separator "Commands:"
      opts.separator "  create\t<account_name> <account_email>\tCreate a new SSH key for a GitHub account"
      opts.separator "  delete\t<account_name>\t\t\tDelete an SSH key for a GitHub account"
      opts.separator "  copy\t\t<account_name>\t\t\tCopy the public key for a GitHub account to the clipboard"
      opts.separator "  use\t\tUse an SSH key for a GitHub account in the current directory"
      opts.separator "  list\t\tList all SSH keys for GitHub accounts"
    end

    opt_parser.order!(args)

    if args.empty?
      puts opt_parser
      exit
    end

    options[:command] = args.shift
    options[:args] = args
    options
  end

  def execute
    begin
      case @command
      when 'create'
        create_account(*@args)
      when 'delete'
        delete_account(get_account_name)
      when 'copy'
        copy_public_key(get_account_name)
      when 'use'
        use_account(get_account_name)
      else
        raise ArgumentError, Localization.get_message("system.incorrect_command").colorize(:color => :red)
      end
    rescue Interrupt
      puts Localization.get_message("system.operation_cancelled").colorize(:color => :red)
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

  def create_account(*args)
    passphrase_option = args.include?('-p')
    args.delete('-p')
    if args.length == 2
      account_name, account_email = args
    else
      puts Localization.get_message("input.account_name")
      account_name = STDIN.gets.chomp
      puts Localization.get_message("input.email")
      account_email = STDIN.gets.chomp
    end
    unless Validation.valid_account_name?(account_name)
      puts Localization.get_message("error.invalid_account_name")
      exit 1
    end

    unless Validation.valid_email?(account_email)
      puts Localization.get_message("error.invalid_email")
      exit 1
    end

    KeyManager.check_key_exists(account_name)

    if passphrase_option
      KeyManager.create_ssh_key(account_name, account_email, true)
    else
      KeyManager.create_ssh_key(account_name, account_email, false)
    end

    KeyManager.add_key_to_agent(account_name)
    KeyManager.add_ssh_config_entry(account_name)

    puts Localization.get_message("ssh.created")
    puts Localization.get_message("ssh.copy_public_key")

    KeyManager.copy_public_key_to_clipboard(account_name)
  end

  def delete_account(*args)
    account_name = args

    if account_name.nil? || account_name.empty?
      puts Localization.get_message("input.account_name")
      account_name = STDIN.gets.chomp
    end

    unless Validation.valid_account_name?(account_name)
      puts Localization.get_message("error.invalid_account_name")
      exit 1
    end

    KeyManager.check_key_exists(account_name)

    puts Localization.get_message("input.confirm_delete")
    confirm_delete = STDIN.gets.chomp

    if confirm_delete == 'y'
      KeyManager.remove_ssh_config_entry(account_name)
      KeyManager.delete(account_name)
      puts Localization.get_message("ssh.deleted")
    else
      puts Localization.get_message("system.operation_cancelled")
    end
  end

  def copy_public_key(*args)
    account_name = args

    if account_name.nil? || account_name.empty?
      puts Localization.get_message("input.account_name")
      account_name = STDIN.gets.chomp
    end

    unless Validation.valid_account_name?(account_name)
      puts Localization.get_message("error.invalid_account_name")
      exit 1
    end

    KeyManager.copy_public_key_to_clipboard(account_name)
  end

  def use_account(name)
    KeyManager.check_key_exists(name)

    ssh_config = File.read(ssh_config_path)
    unless ssh_config.include?("Host github.com-#{name}")
      puts "No matching SSH configuration for '#{name}'."
      exit 1
    end

    unless Dir.exist?(File.join(Dir.pwd, '.git'))
      puts "No git repository found in the current directory. Do you want to initialize a new repository? (Y/n)"
      initialize_repo = STDIN.gets.chomp.downcase
      if initialize_repo == 'y' || initialize_repo == ''
        system("git init")
      else
        puts "Operation cancelled. No git repository initialized."
        return
      end
    end

    puts "Enter new name:"
    new_name = STDIN.gets.chomp
    puts "Enter new email:"
    new_email = STDIN.gets.chomp
    puts "Enter new remote URL:"
    new_url = STDIN.gets.chomp

    system("git config user.name '#{new_name}'")
    system("git config user.email '#{new_email}'")
    system("git remote set-url origin '#{new_url}'")

    puts "Git configuration updated with new name, email, and remote URL."
  end

  def get_account_details
    account_name = InputManager.get_valid_input("input.account_name", :valid_account_name?)
    account_email = InputManager.get_valid_input("input.email", :valid_email?)
    [account_name, account_email]
  end

  def get_account_name
    InputManager.get_valid_input("input.account_name", :valid_account_name?)
  end

  def load_config
    {
      ssh_dir_path: "#{ENV['HOME']}/.ssh",
      ssh_config_path: "#{ENV['HOME']}/.ssh/config",
    }
  end

  def self.run(args)
    options = parse_args(args)
    multigit = MultiGit.new(options[:command], options[:args])
    multigit.execute
  end
end

if __FILE__ == $0
  MultiGit.run(ARGV)
end