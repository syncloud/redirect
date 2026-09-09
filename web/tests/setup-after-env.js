import { config } from '@vue/test-utils'
import i18n from '../src/i18n'

config.global.plugins = [...(config.global.plugins || []), i18n]

jest.setTimeout(30000)
