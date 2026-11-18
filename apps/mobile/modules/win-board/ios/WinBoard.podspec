Pod::Spec.new do |s|
  s.name           = 'WinBoard'
  s.version        = '1.0.0'
  s.summary        = 'Starts, updates and ends the Freebies Win Board Live Activity'
  s.description    = 'Local Expo module for the Win Board Live Activity (see targets/widget/winboard.swift).'
  s.author         = ''
  s.homepage       = 'https://github.com/bitsugar-io/freebies'
  s.license        = 'MIT'
  s.platform       = :ios, '16.4'
  s.source         = { git: '' }
  s.static_framework = true

  s.dependency 'ExpoModulesCore'

  s.pod_target_xcconfig = {
    'DEFINES_MODULE' => 'YES',
    'SWIFT_COMPILATION_MODE' => 'wholemodule'
  }

  s.source_files = "**/*.{h,m,mm,swift}"
end
