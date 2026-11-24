require 'date'
require 'fileutils'
require 'open3'
require 'tempfile'
require 'yaml'

abort '用法: ruby server/scripts/export_week_data.rb [YYYY-MM-DD]' if ARGV.size > 1
now = Time.now.getlocal('+08:00')
begin
  last_day = ARGV.empty? ? now.to_date : Date.iso8601(ARGV.first)
rescue Date::Error
  abort '截止日期格式为 YYYY-MM-DD'
end
first_day = last_day - 6
project = File.expand_path('../..', __dir__)
config = YAML.safe_load(File.read(File.join(project, 'server/config/config.yaml'))).fetch('mysql')
abort '当前配置未连接 energytest' unless config.fetch('database') == 'energytest'
mysql = ['mysql', '--protocol=TCP', '--host', config.fetch('host'), '--port', config.fetch('port').to_s,
         '--user', config.fetch('username'), '--database', 'energytest', '--default-character-set=utf8mb4',
         '--batch', '--quick', '--raw', '--skip-column-names']
env = { 'MYSQL_PWD' => config.fetch('password') }

metadata, error, status = Open3.capture3(env, *mysql, '-e', <<~SQL)
  SELECT TABLE_NAME, COLUMN_NAME, DATA_TYPE FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() ORDER BY TABLE_NAME, ORDINAL_POSITION
SQL
abort error unless status.success?
columns = metadata.lines.map { |line| line.chomp.split("\t") }.group_by(&:first)
date_columns = {
  'user' => 'created_at', 'roles' => 'created_at', 'sys_role' => 'create_time',
  'sys_menu_role' => 'create_time', 'energy' => 'DATETIME', 'summary' => 'date',
  'jy_data' => 'NOEDATE', 'image' => 'uploadTime', 'stocks' => 'time',
  'market_review_days' => 'date', 'market_review_notes' => 'date',
  'trade_daily_summary' => 'trade_date', 'xq_theme_content' => 'created_at',
  'topic_research_runs' => 'created_at', 'topic_research_skills' => 'created_at'
}
window = lambda { |column| "`#{column}` >= '#{first_day}' AND `#{column}` < '#{last_day + 1}'" }
filters = date_columns.transform_values(&window)
%w[market_review_members market_review_studies].each do |table|
  filters[table] = "day_id IN (SELECT id FROM market_review_days WHERE #{window.call('date')})"
end
filters['test_jy'] = %w[jy_buy_date jy_del_date].map do |column|
  "(`#{column}` REGEXP '^[0-9]{4}-[0-9]{2}-[0-9]{2}' AND #{window.call(column)})"
end.join(' OR ')

# Use hex literals for arbitrary strings and blobs so every row occupies one client output line.
def export_literal(column, type)
  name = "`#{column.gsub('`', '``')}`"
  value = case type
          when /\A(?:bigint|int|smallint|tinyint|mediumint|float|double|decimal)\z/ then "CAST(#{name} AS CHAR)"
          when /\A(?:date|datetime|time|timestamp)\z/ then "QUOTE(#{name})"
          when /blob|binary/ then %Q{CONCAT("X'",HEX(#{name}),"'")}
          when /char|text/ then %Q{CONCAT("X'",HEX(CONVERT(#{name} USING utf8mb4)),"'")}
          else abort "不支持导出字段类型 #{column} #{type}"
          end
  "IF(#{name} IS NULL,'NULL',#{value})"
end

queries = ["SET SESSION sql_mode='';", "SET SESSION time_zone='+08:00';",
           'SET SESSION TRANSACTION ISOLATION LEVEL REPEATABLE READ;',
           'START TRANSACTION WITH CONSISTENT SNAPSHOT, READ ONLY;']
prefixes = {}
filters.each do |table, predicate|
  fields = columns.fetch(table)
  literals = fields.map { |_, column, type| export_literal(column, type) }
  names = fields.map { |_, column, _| "`#{column.gsub('`', '``')}`" }
  prefixes[table] = "INSERT INTO `#{table}` (#{names.join(',')}) VALUES"
  queries << "SELECT CONCAT('@#{table}:',COUNT(*)) FROM `#{table}` WHERE #{predicate};"
  queries << "SELECT CONCAT('(', CONCAT_WS(',',#{literals.join(',')}), ')') FROM `#{table}` WHERE #{predicate};"
end
queries << 'COMMIT;'
directory = File.join(project, 'tmp/database-exports')
FileUtils.mkdir_p(directory, mode: 0700)
path = File.join(directory, "energytest_#{first_day}_#{last_day}_#{now.strftime('%H%M%S%6N')}.sql")
counts = Hash.new(0)
expected = {}
Tempfile.create(['.export-', '.partial'], directory) do |file|
  file.puts "-- energytest | Asia/Shanghai | #{first_day} .. #{last_day}",
            'SET NAMES utf8mb4;', 'SET @week_old_tz=@@session.time_zone;',
            "SET SESSION time_zone='+08:00';", 'START TRANSACTION;'
  Open3.popen3(env, *mysql) do |input, output, errors, process|
    error_reader = Thread.new { errors.read }
    # Supply SQL without blocking on the client's output pipe.
    writer = Thread.new do
      begin
        input.puts(queries)
      ensure
        input.close
      end
    end
    table = nil
    output.each_line do |line|
      line.chomp!
      if line.start_with?('@')
        table, count = line[1..-1].split(':')
        expected[table] = Integer(count)
      else
        raise '数据行不完整' unless line.start_with?('(') && line.end_with?(')')
        file.puts "#{prefixes.fetch(table)} #{line};"
        counts[table] += 1
      end
    end
    writer.value
    abort error_reader.value unless process.value.success?
  end
  abort '未通过逐表行数校验' unless expected.keys == filters.keys && expected.all? { |t, n| counts[t] == n }
  file.puts 'COMMIT;', 'SET SESSION time_zone=@week_old_tz;'
  file.close
  File.link(file.path, path)
end
puts "已导出 #{first_day} 至 #{last_day} 共 #{counts.values.sum} 条：#{path}"
