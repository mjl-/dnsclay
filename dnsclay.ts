let rootElem: HTMLElement
let crumbElem = dom.span()
let pageElem = dom.div(style({padding: '1em'}), dom.div(style({textAlign: 'center'}), 'Loading...'))

let version: string
let dnsTypeNames: { [key: number]: string } = {}

const client = new api.Client()

const link = (href: string, anchor: string) => dom.a(attr.href(href), anchor)

const trimDot = (s: string) => s.replace(/\.$/, '')

const check = async <T>(elem: {disabled: boolean}, fn: () => Promise<T>): Promise<T> => {
	const overlay = dom.div(style({position: 'fixed', top: 0, left: 0, right: 0, bottom: 0, zIndex: 2, backgroundColor: '#ffffff00'}))
	document.body.append(overlay)
	pageElem.classList.toggle('loading', true)
	if (elem) {
		elem.disabled = true
	}

	try {
		const r = await fn()
		return r
	} catch (err: any) {
		alert('Error: '+err.message)
		throw err
	} finally {
		overlay.remove()
		pageElem.classList.toggle('loading', false)
		if (elem) {
			elem.disabled = false
		}
	}
}

const popupOpts = (opaque: boolean, ...kids: ElemArg[]) => {
	const origFocus = document.activeElement
	const close = () => {
		if (!root.parentNode) {
			return
		}
		root.remove()
		if (origFocus && origFocus instanceof HTMLElement && origFocus.parentNode) {
			origFocus.focus()
		}
	}
	let content: HTMLElement
	const root = dom.div(
		style({position: 'fixed', top: 0, right: 0, bottom: 0, left: 0, paddingTop: '5vh', backgroundColor: opaque ? '#ffffff' : 'rgba(0, 0, 0, 0.1)', display: 'flex', alignItems: 'flex-start', justifyContent: 'center', zIndex: opaque ? 3 : 1}),
		opaque ? [] : [
			function keydown(e: KeyboardEvent) {
				if (e.key === 'Escape') {
					e.stopPropagation()
					close()
				}
			},
			function click(e: MouseEvent) {
				e.stopPropagation()
				close()
			},
		],
		content=dom.div(
			attr.tabindex('0'),
			style({backgroundColor: 'white', borderRadius: '.25em', padding: '1em', boxShadow: '0 0 20px rgba(0, 0, 0, 0.1)', border: '1px solid #ddd', maxWidth: '95vw', overflowX: 'auto', maxHeight: '95vh', overflowY: 'auto'}),
			function click(e: MouseEvent) {
				e.stopPropagation()
			},
			kids,
		)
	)
	document.body.appendChild(root)
	content.focus()
	return close
}

const trimPrefix = (s: string, prefix: string) => s.startsWith(prefix) ? s.substring(prefix.length) : s
const trimSuffix = (s: string, suffix: string) => s.endsWith(suffix) ? s.substring(0, s.length-suffix.length) : s

const chunked = <T>(l: T[], len: number): T[][] => {
	const r: T[][] = []
	while (l.length > 0) {
		r.push(l.slice(0, len))
		l = l.slice(len)
	}
	return r
}

const popup = (...kids: ElemArg[]) => popupOpts(false, ...kids)

const availableProviders = async (): Promise<[Map<string, api.sherpadocStrings>, api.sherpadocStruct[]]> => {
	const docs = await client.Docs()
	const stringEnums = new Map<string, api.sherpadocStrings>()
	for (const e of (docs.Strings || [])) {
		stringEnums.set(e.Name, e)
	}
	const providers = (docs.Structs || []).filter(struct => struct.Name.startsWith('Provider_'))
	return [stringEnums, providers]
}

interface ProviderConfigField {
	elem: HTMLInputElement | HTMLSelectElement
	nullable: boolean
	type: 'string' | 'number' | 'bool'
}

interface ProviderFields {
	fieldMap: Map<string, ProviderConfigField>
	root: HTMLElement
}

const providerConfigJSON = (fields: ProviderFields) => {
	type FieldValue = string | boolean | number | null
	const config: { [key: string]: FieldValue } = {}
	for (const [k, f] of fields.fieldMap) {
		let v: FieldValue = null
		if (f.nullable && !f.elem.value) {
		} else if (f.type === 'bool') {
			v = (f.elem as HTMLInputElement).checked
		} else if (f.type === 'number') {
			v = parseInt(f.elem.value)
		} else {
			v = f.elem.value
		}
		config[k] = v
	}
	return JSON.stringify(config)
}

const providerFields = (p: api.sherpadocStruct, stringEnums: Map<string, api.sherpadocStrings>, configJSON: string | null): ProviderFields => {
	const fieldMap = new Map<string, ProviderConfigField>()

	let config: { [ key: string]: any } = {}
	if (configJSON) {
		config = JSON.parse(configJSON)
	}

	const root = dom.div(
		(p.Fields || []).map(f => {
			let tw = f.Typewords || []
			let nullable = false
			if (!tw[0]) {
				alert('missing type word')
				throw new Error('missing type word')
			}
			if (tw[0] === 'nullable') {
				tw = tw.slice(1)
			}
			if (!tw[0]) {
				alert('missing type word')
				throw new Error('missing type word')
			}

			if (tw[0] === 'bool') {
				const input = dom.input(attr.type('checkbox'), config[f.Name] === true ? attr.checked('') : [])
				fieldMap.set(f.Name, {elem: input, nullable: nullable, type: 'bool'})
				return dom.div(
					dom.div(style({margin: '.5ex 0'}), dom.label(input, ' ', '"'+f.Name+'"')),
					f.Docs ? dom.div(style({fontStyle: 'italic', maxWidth: '40em', marginBottom: '2ex'}), '"'+f.Docs+'"') : [],
				)
			}

			let input: HTMLInputElement | HTMLSelectElement = dom.input()
			let typ: 'string' | 'number' = 'string'
			if (tw[0] === 'int32' || tw[0] === 'int64' || tw[0] === 'uint32' || tw[1] === 'uint64') {
				input = dom.input(attr.type('number'), config[f.Name] ? attr.value(''+config[f.Name]) : [])
				typ = 'number'
			} else if (tw[0] === 'string') {
				input = dom.input(config[f.Name] ? attr.value(''+config[f.Name]) : [])
			} else {
				const values = stringEnums.get(tw[0])
				if (values) {
					input = dom.select(
						(values.Values || []).map(v =>
							dom.option(`${v.Name} - ${v.Value}`, attr.value(v.Value), config[f.Name] === v.Value ? attr.selected('') : [])
						),
					)
				} else {
					alert(`unknown type "${tw.join(' ')}" for field ${f.Name}`)
					input = dom.input()
				}
			}
			fieldMap.set(f.Name, {elem: input, nullable: nullable, type: typ})
			return dom.div(
				dom.div(dom.label('"'+f.Name+'"')),
				dom.div(style({margin: '.5ex 0'}), input),
				f.Docs ? dom.div(style({fontStyle: 'italic', maxWidth: '40em', marginBottom: '2ex'}), '"'+f.Docs+'"') : [],
			)
		})
	)
	return {root: root, fieldMap: fieldMap}
}

const pageHome = async () => {
	let [zones0] = await Promise.all([
		client.Zones(),
	])
	let zones = zones0 || []

	dom._kids(crumbElem,
		dom.a(attr.href('#'), 'Home'),
	)
	document.title = 'Dnsclay'

	let zonesTbody: HTMLElement

	const root = dom.div(
		dom.div(
			dom.clickbutton('Add zone', async function click() {
				let zone: HTMLInputElement
				let fieldset: HTMLFieldSetElement
				let testResult: HTMLElement
				let newProviderConfigName: HTMLInputElement
				let existingProviderConfigName: HTMLSelectElement

				const [stringEnums, providers] = await availableProviders()
				const providerConfigs = await client.ProviderConfigs() || []

				let fields: ProviderFields
				let providerName = ''

				const updateProviderConfig = () => {
					providerName = (fieldset!.querySelector('input[name=provider]:checked') as HTMLInputElement).value
					const p = providers.find(p => p.Name === 'Provider_'+providerName)
					if (!p) {
						alert('cannot find provider '+providerName)
						return
					}

					dom._kids(providerConfigBox,
						style({display: 'flex', flexDirection: 'column', gap: '2ex'}),
						dom.label(
							dom.div('Name'),
							dom.div(newProviderConfigName=dom.input(attr.required(''), attr.value(newProviderConfigName?.value || zone.value))),
						),
						dom.div(
							style({padding: '1em', border: '1px solid #ddd'}),
							dom.h2('"'+providerName+'" fields'),
							dom.p('Implemented through github/libdns/'+providerName+', see ', dom.a(attr.href('https://pkg.go.dev/github.com/libdns/'+providerName), 'documentation', attr.rel('noreferrer noopener'))),
							dom.div(
								style({display: 'flex', flexDirection: 'column', gap: '2ex'}),
								fields=providerFields(p, stringEnums, null),
							)
						),
					)
				}
				let providerConfigBox: HTMLElement

				const close = popup(
					dom.div(
						dom.h1('New zone'),
						dom.form(
							async function submit(e: SubmitEvent) {
								e.preventDefault()
								e.stopPropagation()

								// Test config.
								if (testResult) {
									testResult.innerText = ''
								}
								if (!zone.value) {
									alert('Zone required.')
									return
								}

								let pName = ''
								let pcJSON = ''
								if (existingProviderConfigName.value) {
									const pc = providerConfigs.find(pc => pc.Name === existingProviderConfigName.value)
									if (!pc) {
										alert('Provider config not found.')
										return
									}
									pName = pc.ProviderName
									pcJSON = pc.ProviderConfigJSON
								} else {
									if (!fields) {
										alert('No provider selected.')
										return
									}
									pName = providerName
									pcJSON = providerConfigJSON(fields)
								}

								const nrecords = await check(fieldset, () => client.ProviderConfigTest(trimSuffix(zone.value, '.')+'.', pName, pcJSON))
								testResult.innerText = 'Success, found '+nrecords+' DNS records'
							},
							fieldset=dom.fieldset(
								style({display: 'flex', flexDirection: 'column', gap: '2ex'}),
								dom.div(
									dom.div(dom.label('Zone')),
									zone=dom.input(attr.required('')),
								),
								dom.div(
									dom.div(dom.label('Create new provider config')),
									dom.div(
										style({display: 'flex', gap: '1em'}),
										chunked(providers, 10).map(plist =>
											dom.div(
												plist.map(p => {
													return dom.div(
														dom.label(
															dom.input(attr.type('radio'), attr.name('provider'), attr.value(trimPrefix(p.Name, 'Provider_')), function change() { updateProviderConfig() }),
															' ', trimPrefix(p.Name, 'Provider_')
														),
													)
												})
											)
										)
									)
								),
								providerConfigBox=dom.div(),
								dom.label(
									dom.div('Use existing provider config'),
									dom.div(
										existingProviderConfigName=dom.select(
											dom.option('', attr.value('')),
											providerConfigs.map(pc => dom.option(pc.Name)),
										),
									),
								),
								dom.div(
									dom.submitbutton('Test config'), ' ',
									testResult=dom.span(),
								),
								dom.div(
									dom.clickbutton('Add zone', async function click() {
										let pcName = existingProviderConfigName.value
										if (!pcName) {
											if (!fields) {
												alert('No provider selected.')
												return
											}
											let pc: api.ProviderConfig = {
												Name: newProviderConfigName.value,
												ProviderName: providerName,
												ProviderConfigJSON: providerConfigJSON(fields),
											}
											pc = await check(fieldset, () => client.ProviderConfigAdd(pc))
											pcName = pc.Name
										}

										const z: api.Zone = {
											Name: trimSuffix(zone.value, '.')+'.',
											ProviderConfigName: pcName,
											SerialLocal: 0,
											SerialRemote: 0,
											SyncInterval: 0,
											RefreshInterval: 0,
											NextSync: new Date(),
											NextRefresh: new Date(),
										}
										const nz = await check(fieldset, () => client.ZoneAdd(z, [])) // todo: allow specifying notifies
										zones.push(nz)
										render()
										close()
									}),
								),
							),
						),
					),
				)
				zone.focus()
			}),
		),
		dom.br(),
		dom.h1('Zones (Domains)'),
		dom.table(
			dom.thead(
				dom.tr(
					dom.th('Zone'),
					dom.th('Provider Config'),
					dom.th('Last sync'),
					dom.th('Last record change'),
					dom.th('Serial'),
					dom.th('Refresh next/interval'),
					dom.th('Sync next/interval'),
				),
			),
			zonesTbody=dom.tbody(),
		),
	)

	const render = () => {
		const now = new Date()
		dom._kids(zonesTbody,
			zones.length ? [] : dom.tr(dom.td(attr.colspan('6'), 'No zones.', style({textAlign: 'left'}))),
			zones.map(z =>
				dom.tr(
					dom.td(dom.a(attr.href('#zones/'+trimDot(z.Name)), trimDot(z.Name))),
					dom.td(z.ProviderConfigName),
					dom.td(z.LastSync ? [formatAge(z.LastSync), attr.title(formatDate(z.LastSync))] : []),
					dom.td(z.LastRecordChange ? [formatAge(z.LastRecordChange), attr.title(formatDate(z.LastRecordChange))] : []),
					dom.td(
						''+z.SerialLocal,
						z.SerialLocal !== z.SerialRemote ? ' (at remote: '+z.SerialRemote+')' : '',
						attr.title(`Next SOA check in ${ formatAge(undefined, z.NextRefresh) } at ${ formatDate(z.NextRefresh) }.\nNext sync in ${ formatAge(undefined, z.NextSync) } at ${ formatDate(z.NextSync) }.`),
					),
					dom.td(
						formatAge(now, z.NextRefresh),
						' / ',
						formatAge(now, new Date(now.getTime() + z.RefreshInterval/(1000*1000))),
					),
					dom.td(
						formatAge(now, z.NextSync),
						' / ',
						formatAge(now, new Date(now.getTime() + z.SyncInterval/(1000*1000))),
					),
				)
			),
		)
	}
	render()

	return root
}

const formatAge = (start?: Date, end?: Date) => {
	const second = 1
	const minute = 60*second
	const hour = 60*minute
	const day = 24*hour
	const week = 7*day
	const year = 365*day
	const periods = [year, week, day, hour, minute, second]
	const suffix = ['y', 'w', 'd', 'h', 'm', 's']

	if (!start) {
		start = new Date()
	}
	if (!end) {
		end = new Date()
	}
	let e = end.getTime()/1000
	let t = e - start.getTime()/1000
	let s = ''
	for (let i = 0; i < periods.length; i++) {
		const p = periods[i]
		if (t >= 2*p || i === periods.length-1) {
			const n = Math.round(t/p)
			s = '' + n + suffix[i]
			break
		}
	}
	return s
}

const formatDate = (dt: Date) => {
	return new Intl.DateTimeFormat(undefined, {
		weekday: 'short',
		year: 'numeric',
		month: 'short',
		day: 'numeric',
		hour: 'numeric',
		minute: 'numeric',
	}).format(dt)
}

const pageZone = async (zonestr: string) => {
	let [zone, providerConfig, notifies0, credentials0, records0] = await client.Zone(zonestr+'.')
	let notifies = notifies0 || []
	let credentials = credentials0 || []
	let records = records0 || []

	dom._kids(crumbElem,
		dom.a(attr.href('#'), 'Home'), ' / ',
		dom.a(attr.href('#zones/'+trimDot(zone.Name)), 'Zone '+trimDot(zone.Name)),
	)
	document.title = 'Dnsclay - Zone '+trimDot(zone.Name)

	const shortname = (s: string) => {
		s = s.substring(0, s.length-zone.Name.length)
		if (s) {
			s = s.substring(0, s.length-1)
		}
		return s
	}

	const age = (r: api.Record) => {
		let s = ''
		let title = 'First: ' + formatDate(r.First) + ' (serial ' + r.SerialFirst + ')\n'
		if (r.Deleted) {
			s = formatAge(r.Deleted)
			title += 'Deleted: ' + formatDate(r.Deleted) + ' (serial ' + r.SerialDeleted + ')\n'
		} else {
			s = formatAge(r.First)
		}
		return dom.span(s, attr.title(title))
	}

	let showHistoric: HTMLInputElement
	let showDNSSEC: HTMLInputElement
	let recordsTbody: HTMLElement

	const refresh = async (elem: {disabled: boolean}) => {
		const [nzone, npc, nnotifies, ncredentials, nrecords] = await check(elem, () => client.Zone(zone.Name))
		zone = nzone
		providerConfig = npc
		notifies = nnotifies || []
		credentials = ncredentials || []
		records = nrecords || []
		render()
	}

	const root = dom.div(
		dom.div(
			dom.p(`Provider config: ${ zone.ProviderConfigName } (provider ${ providerConfig.ProviderName })`),
			dom.clickbutton(
				'Edit zone config',
				async function click(e: {target: HTMLButtonElement}) {
					let fieldset: HTMLFieldSetElement
					let refreshival: HTMLInputElement
					let syncival: HTMLInputElement
					let providerConfigName: HTMLSelectElement

					const providerConfigs = await check(e.target, () => client.ProviderConfigs()) || []

					const close = popup(
						dom.h1('Edit zone'),
						dom.br(),
						dom.form(
							async function submit(e: SubmitEvent) {
								e.preventDefault()
								e.stopPropagation()

								const nz = {...zone}
								nz.ProviderConfigName = providerConfigName.value
								nz.RefreshInterval = 1000*1000*1000 * parseInt(refreshival.value)
								nz.SyncInterval = 1000*1000*1000 * parseInt(syncival.value)
								zone = await check(fieldset, () => client.ZoneUpdate(nz))
								close()
							},
							fieldset=dom.fieldset(
								style({display: 'flex', flexDirection: 'column', gap: '2ex'}),
								dom.label(
									dom.div('Refresh interval (in seconds)'),
									refreshival=dom.input(attr.type('number'), attr.required(''), attr.value(''+(zone.RefreshInterval/(1000*1000*1000)))),
								),
								dom.label(
									dom.div('Sync interval (in seconds)'),
									syncival=dom.input(attr.type('number'), attr.required(''), attr.value(''+(zone.SyncInterval/(1000*1000*1000)))),
								),
								dom.label(
									dom.div('Provider config'),
									providerConfigName=dom.select(
										providerConfigs.sort((a, b) => a.Name < b.Name ? -1 : 1).map(pc => dom.option(pc.Name)),
										prop({value: zone.ProviderConfigName}),
									),
								),
								dom.div(
									dom.submitbutton('Save')
								),
							),
						),
					)
				},
			), ' ',
			dom.clickbutton(
				'Edit provider config',
				async function click(e: {target: HTMLButtonElement}) {
					let fieldset: HTMLFieldSetElement

					const [stringEnums, providers] = await check(e.target, () => availableProviders())
					const p = providers.find(p => p.Name === 'Provider_'+providerConfig.ProviderName)
					if (!p) {
						alert('cannot find provider '+providerConfig.ProviderName)
						return
					}

					let testResult: HTMLElement
					let fields: ProviderFields

					const close = popup(
						dom.h1('Edit provider config'),
						dom.form(
							async function submit(e: SubmitEvent) {
								e.preventDefault()
								e.stopPropagation()

								// Test config.
								testResult.innerText = ''
								const nrecords = await check(fieldset, () => client.ProviderConfigTest(zone.Name, providerConfig.ProviderName, providerConfigJSON(fields)))
								testResult.innerText = 'Success, found '+nrecords+' DNS records'
							},
							fieldset=dom.fieldset(
								style({display: 'flex', flexDirection: 'column', gap: '2ex'}),
								dom.label(
									dom.div('Name'),
									dom.div(dom.input(attr.value(providerConfig.Name), attr.disabled(''))),
								),
								dom.label(
									dom.div('Provider'),
									dom.div(
										dom.select(
											attr.disabled(''),
											dom.option(providerConfig.ProviderName),
											prop({value: providerConfig.ProviderName}),
										),
									),
								),
								dom.div(
									style({padding: '1em', border: '1px solid #ddd'}),
									dom.h2('Provider config'),
									dom.div(
										style({display: 'flex', flexDirection: 'column', gap: '2ex'}),
										fields=providerFields(p, stringEnums, providerConfig.ProviderConfigJSON),
									),
								),
								dom.div(
									dom.submitbutton('Test config'), ' ',
									testResult=dom.span(),
								),
								dom.div(
									dom.clickbutton('Save', async function click() {
										let npc: api.ProviderConfig = {
											Name: providerConfig.Name, // todo: allow editing, need to rename it for all users in database.
											ProviderName: providerConfig.ProviderName, // todo: allow changing too
											ProviderConfigJSON: providerConfigJSON(fields),
										}
										providerConfig = await check(fieldset, () => client.ProviderConfigUpdate(npc))
										close()
									}),
								),
							),
						),
					)
				},
			),
		),
		dom.br(),

		dom.div(
			style({display: 'flex', gap: '1em'}),
			dom.div(
				style({backgroundColor: '#f4f4f4', border: '1px solid #ddd', borderRadius: '.25em', padding: '.5em'}),
				dom.div(
					style({display: 'flex', gap: '.5em', alignItems: 'baseline'}),
					dom.h2(
						'DNS NOTIFY addresses',
					),
					dom.clickbutton('Add', function click() {
						let address: HTMLInputElement
						let fieldset: HTMLFieldSetElement

						const close = popup(
							dom.h1('Add DNS NOTIFY address'),
							dom.form(
								async function submit(e: SubmitEvent) {
									e.preventDefault()
									e.stopPropagation()
									let zn: api.ZoneNotify = {
										ID: 0,
										Created: new Date(),
										Zone: zone.Name,
										Protocol: (fieldset.querySelector('input[name=notifyprotocol]:checked') as HTMLInputElement)?.value || '',
										Address: address.value,
									}
									const nzn = await check(fieldset, () => client.ZoneNotifyAdd(zn))
									notifies.push(nzn)
									close()
									location.reload() // todo: render the list again
								},
								fieldset=dom.fieldset(
									style({display: 'flex', flexDirection: 'column', gap: '2ex'}),
									dom.div(
										dom.div(dom.label('Protocol')),
										dom.label(dom.input(attr.type('radio'), attr.name('notifyprotocol'), attr.value('tcp')), ' tcp'), ' ',
										dom.label(dom.input(attr.type('radio'), attr.name('notifyprotocol'), attr.value('udp')), ' udp'),
									),
									dom.div(
										dom.div(dom.label('Address')),
										address=dom.input(attr.type('required'), attr.placeholder('127.0.0.1:53')),
									),
									dom.div(
										dom.submitbutton('Add'),
									),
								),
							),
						)
					}),
				),
				dom.table(
					dom.thead(
						dom.tr(
							dom.th('Protocol'),
							dom.th('Address'),
							dom.th(),
						),
					),
					dom.tbody(
						notifies.length ? [] : dom.tr(dom.td(attr.colspan('3'), 'No notify addressses.', style({textAlign: 'left'}))),
						notifies.map(n => {
							const row = dom.tr(
								dom.td(n.Protocol),
								dom.td(n.Address),
								dom.td(
									dom.clickbutton('Notify', async function click(e: {target: HTMLButtonElement}) {
										await check(e.target, () => client.ZoneNotify(n.ID))
									}), ' ',
									dom.clickbutton('Delete', async function click(e: {target: HTMLButtonElement}) {
										if (!confirm('Are you sure?')) {
											return
										}
										await check(e.target, () => client.ZoneNotifyDelete(n.ID))
										notifies.splice(notifies.indexOf(n), 1)
										row.remove()
									}),
								),
							)
							return row
						}),
					),
				),
			),

			dom.div(
				style({backgroundColor: '#f4f4f4', border: '1px solid #ddd', borderRadius: '.25em', padding: '.5em'}),
				dom.div(
					style({display: 'flex', gap: '.5em', alignItems: 'baseline'}),
					dom.h2(
						'Credentials',
					), ' ',
					dom.clickbutton('Add', function click() {
						let name: HTMLInputElement
						let key: HTMLInputElement
						let fieldset: HTMLFieldSetElement

						const close = popup(
							dom.h1('Add credential'),
							dom.p('For use with DNS UPDATE and DNS AXFR/IXFR.'),
							dom.form(
								async function submit(e: SubmitEvent) {
									e.preventDefault()
									e.stopPropagation()
									const typ = (fieldset.querySelector('input[name=credentialtype]:checked') as HTMLInputElement)?.value || ''
									let c: api.Credential = {
										ID: 0,
										Created: new Date(),
										Name: name.value,
										Type: typ,
										TSIGSecret: typ === 'tsig' ? key.value : '',
										TLSPublicKey: typ === 'tlspubkey' ? key.value : '',
									}
									const nc = await check(fieldset, () => client.ZoneCredentialAdd(zone.Name, c))
									credentials.push(nc)
									close()
									location.reload() // todo: render the list again
								},
								fieldset=dom.fieldset(
									style({display: 'flex', flexDirection: 'column', gap: '2ex'}),
									dom.div(
										dom.div(dom.label('Name')),
										name=dom.input(attr.type('required'), attr.placeholder('name-with-dashes-or-dots'), style({width: '100%'})),
										dom.div(style({fontStyle: 'italic'}), 'Must be a valid DNS name for TSIG.'),
									),
									dom.div(
										dom.div(dom.label('Type')),
										dom.label(dom.input(attr.type('radio'), attr.name('credentialtype'), attr.value('tsig')), ' TSIG'), ' ',
										dom.label(dom.input(attr.type('radio'), attr.name('credentialtype'), attr.value('tlspubkey')), ' TLS public key'),
									),
									dom.div(
										dom.div(dom.label('TSIG secret or TLS public key')),
										key=dom.input(style({width: '100%'})),
										dom.div(style({fontStyle: 'italic'}), 'In case of a TSIG secret, if left empty, a random key will be generated.'),
									),
									dom.div(
										dom.submitbutton('Add'),
									),
								),
							),
						)
					}),
				),
				dom.table(
					dom.thead(
						dom.tr(
							dom.th('Name'),
							dom.th('Type'),
							dom.th('TSIG Secret / TLS public key'),
							dom.th('Age'),
							dom.th(''),
						),
					),
					dom.tbody(
						credentials.length ? [] : dom.tr(dom.td(attr.colspan('5'), 'No credentials.', style({textAlign: 'left'}))),
						credentials.map(c => {
							const row = dom.tr(
								dom.td(c.Name),
								dom.td(c.Type),
								dom.td(c.Type === 'tsig' ?
									dom.clickbutton('Show', function click(e: {target: HTMLButtonElement}) {
										e.target.replaceWith(dom.span(c.TSIGSecret))
									}) : c.TLSPublicKey,
								),
								dom.td(formatAge(c.Created), attr.title(formatDate(c.Created))),
								dom.td(
									dom.clickbutton('Delete', async function click(e: {target: HTMLButtonElement}) {
										if (!confirm('Are you sure?')) {
											return
										}
										await check(e.target, () => client.ZoneCredentialDelete(c.ID))
										credentials.splice(credentials.indexOf(c), 1)
										row.remove()
									}),
								),
							)
							return row
						}),
					),
				),
			),
		),
		dom.br(),

		dom.div(
			style({display: 'flex', gap: '.5em', alignItems: 'baseline'}),
			dom.h2('Records'), ' ',
			dom.clickbutton('Fetch latest records', async function click(e: {target: HTMLButtonElement}) {
				const [_, nrecords] = await check(e.target, () => client.ZoneRefresh(zone.Name))
				records = nrecords || []
				render()
			}), ' ',
			dom.clickbutton('Add record', function click() {
				let xtype: HTMLSelectElement
				let relName: HTMLInputElement
				let ttl: HTMLInputElement
				let value: HTMLInputElement
				let fieldset: HTMLFieldSetElement

				// Meta types that we don't let the user create.
				const skipTypes = ['Reserved', 'None', 'NXNAME', 'OPT', 'UINFO', 'UID', 'GID', 'UNSPEC', 'TKEY', 'TSIG', 'IXFR', 'AXFR', 'MAILA', 'MAILB', 'ANY']
				// Types we show first in the list. There is a long tail of uninteresting records.
				const firstTypes = ['A', 'AAAA', 'CAA', 'CNAME', 'DNSKEY', 'DS', 'MX', 'NS', 'OPENPGPKEY', 'PTR', 'SMIMEA', 'SOA', 'SRV', 'SSHFP', 'SVCB', 'TLSA', 'TXT']

				const close = popup(
					dom.h1('New record'),
					dom.form(
						async function submit(e: SubmitEvent) {
							e.preventDefault()
							e.stopPropagation()

							const nr: api.RecordNew = {
								RelName: relName.value,
								Type: parseInt(xtype.value),
								TTL: parseInt(ttl.value),
								Value: value.value,
							}
							const r = await check(fieldset, () => client.RecordAdd(zone.Name, nr))
							records.push(r)
							render()
							await refresh(fieldset)
							close()
						},
						fieldset=dom.fieldset(
							style({display: 'flex', flexDirection: 'column', gap: '2ex'}),
							dom.div(
								dom.label(
									dom.div('Type'),
									xtype=dom.select(
										dom.optgroup(
											attr.label('Common'),
											Object.entries(dnsTypeNames).filter(t => firstTypes.includes(t[1])).sort((ta, tb) => firstTypes.indexOf(ta[1]) - firstTypes.indexOf(tb[1])).map(t => dom.option(attr.value(t[0]), t[1])),
										),
										dom.optgroup(
											attr.label('Others'),
											Object.entries(dnsTypeNames).filter(t => !firstTypes.includes(t[1]) && !skipTypes.includes(t[1])).sort((ta, tb) => ta[1] < tb[1] ? -1 : 1).map(t => dom.option(attr.value(t[0]), t[1])),
										),
									),
								),
							),
							dom.div(
								dom.label(
									dom.div('Name'),
									relName=dom.input(), '.'+zone.Name,
								),
							),
							dom.div(
								dom.label(
									dom.div('TTL'),
									ttl=dom.input(attr.type('number'), attr.required(''), attr.value('300')),
								),
							),
							dom.div(
								dom.label(
									dom.div('Value'),
									value=dom.input(style({width: '60em'})),
								),
							),
							dom.div(
								dom.submitbutton('Add record'),
							),
						),
					),
				)
				xtype.focus()
			}), ' ',
			dom.clickbutton('Import records', attr.title('Import records from zone file'), function click() {
				let zonefile: HTMLTextAreaElement
				let fieldset: HTMLFieldSetElement

				const close = popup(
					dom.h1('Import records from zone file'),
					dom.form(
						async function submit(e: SubmitEvent) {
							e.preventDefault()
							e.stopPropagation()

							const rl = await check(fieldset, () => client.ZoneImportRecords(zone.Name, zonefile.value))
							records.push(...(rl || []))
							render()
							await refresh(fieldset)
							close()
						},
						fieldset=dom.fieldset(
							style({display: 'flex', flexDirection: 'column', gap: '2ex'}),
							dom.div(
								dom.label(
									dom.div('Zone file'),
									zonefile=dom.textarea('$TTL 300 ; default 5m\n$ORIGIN '+zone.Name+'\n\n; record syntax: name ttl type value\nshortname 300 A 1.2.3.4\n', style({width: '60em'}), attr.rows('10')),
								),
							),
							dom.div(
								dom.submitbutton('Import'),
							),
						),
					),
				)
				zonefile.focus()
			}), ' ',
			dom.clickbutton('Purge history', attr.title('Remove history with previously existing but now removed records. History is used by IXFR for incremental zone transfers, but IXFR attempts will fall back to AXFR if history is not available.'), async function click(e: {target: HTMLButtonElement}) {
				if (!confirm('Are you sure?')) {
					return
				}
				const [_, nrecords] = await check(e.target, () => client.ZonePurgeHistory(zone.Name))
				records = nrecords || []
				render()
			}), ' ',
		),
		dom.div(
			dom.label(
				showHistoric=dom.input(
					attr.type('checkbox'),
					localStorage.getItem('showHistoric') === 'yes' ? attr.checked('') : [],
					function change() {
						if (showHistoric.checked) {
							localStorage.setItem('showHistoric', 'yes')
						} else {
							localStorage.removeItem('showHistoric')
						}
						render()
					},
				),
				' Show historic records',
			), ' ',
			dom.label(
				showDNSSEC=dom.input(
					attr.type('checkbox'),
					localStorage.getItem('showDNSSEC') === 'yes' ? attr.checked('') : [],
					function change() {
						if (showDNSSEC.checked) {
							localStorage.setItem('showDNSSEC', 'yes')
						} else {
							localStorage.removeItem('showDNSSEC')
						}
						render()
					},
				),
				' Show DNSSEC signature records',
				attr.title('RRSIG, NSEC and NSEC3 records are hidden by default'),
			),
		),
		dom.table(
			dom._class('hover'),
			style({width: '100%'}),
			dom.thead(
				dom.tr(
					dom.th('Age'),
					dom.th('Name'),
					dom.th('Type'),
					dom.th('TTL'),
					dom.th('Value', style({width: '100%'})),
					dom.th('Actions'),
				),
			),
			recordsTbody=dom.tbody(),
		),
		dom.br(),
		dom.h2('Danger'),
		dom.clickbutton('Remove zone', attr.title('Remove zone from management in dnsclay. The zone and its records are not changed at the provider.'), async function click(e: {target: HTMLButtonElement}) {
			if (!confirm('Are you sure you want to remove this zone from management in dnsclay? The zone and its records are not changed at the provider.')) {
				return
			}
			await check(e.target, () => client.ZoneDelete(zone.Name))
			location.hash = '#'
		})
	)

	const render = () => {
		// todo: implement sorting on other columns. at least type.
		records = records.sort((a, b) => {
			if (a.AbsName !== b.AbsName) {
				return a.AbsName.split('/').reverse().join('/') < b.AbsName.split('/').reverse().join('/') ? -1 : 1
			}
			const ta = dnsTypeNames[a.Type] || (''+a.Type)
			const tb = dnsTypeNames[b.Type] || (''+b.Type)
			if (ta !== tb) {
				return ta < tb ? -1 : 1
			}
			const da = !!a.Deleted
			const db = !!b.Deleted
			if (da !== db) {
				return da ? 1 : -1
			}
			if (!a.Deleted && a.Value !== b.Value) {
				return a.Value < b.Value ? -1 : 1
			}
			let tma = a.Deleted || a.First
			let tmb = b.Deleted || b.First
			return tma.getTime() > tmb.getTime() ? -1 : 1
		})

		// todo: add checkboxes to select multiple/all records and do mass operations like changing TTL and deleting them.
		// todo: add mechanism to show history for an rrset (name + type). possibly a popup

		const typeRRSIG = 46
		const typeNSEC = 47
		const typeNSEC3 = 50

		dom._kids(recordsTbody,
			records.filter(r => (showHistoric.checked || !r.Deleted) && (showDNSSEC.checked || (r.Type !== typeRRSIG && r.Type !== typeNSEC && r.Type !== typeNSEC3))).map((r, index) => {
				const formName = 'form-record-'+index
				let ttl: HTMLInputElement
				let value: HTMLInputElement
				let fieldset: HTMLFieldSetElement

				const removedInputStyle = r.Deleted ? style({color: '#888', backgroundColor: '#eee'}) : []

				return dom.tr(
					r.Deleted ? [style({color: '#888'}), attr.title('Historic/deleted record')] : [],
					dom.td(age(r), style({color: '#888'})),
					dom.td(
						shortname(r.AbsName), style({textAlign: 'right'}), removedInputStyle,
					),
					dom.td(
						dnsTypeNames[r.Type] || (''+r.Type),
						attr.title('Type '+r.Type + (r.ProviderID ? '\nID at provider: '+r.ProviderID : '')),
						style({textAlign: 'left'}),
					),
					dom.td(
						ttl=dom.input(attr.value(''+r.TTL), attr.type('number'), attr.required(''), attr.form(formName), style({width: '5em', textAlign: 'right'}), removedInputStyle),
					),
					dom.td(
						value=dom.input(attr.value(r.Value), attr.form(formName), style({width: '100%'}), removedInputStyle),
					),
					dom.td(
						style({whiteSpace: 'nowrap'}),
						r.Deleted ?
							dom.form(
								style({display: 'inline'}), // Prevent getting buttons on separate lines.
								attr.id(formName),
								async function submit(e: SubmitEvent) {
									e.preventDefault()
									e.stopPropagation()
									if (!confirm('Are you sure you want to recreate this record?')) {
										return
									}

									const nr: api.RecordNew = {
										RelName: shortname(r.AbsName),
										Type: r.Type,
										TTL: parseInt(ttl.value),
										Value: value.value,
									}
									const xr = await check(fieldset, () => client.RecordAdd(zone.Name, nr))
									records.push(xr)
									render()
									await refresh(fieldset)
								},
								fieldset=dom.fieldset(
									dom.submitbutton('Recreate'),
								),
							) :
							dom.span(
								dom.form(
									style({display: 'inline'}), // Prevent getting buttons on separate lines.
									attr.id(formName),
									async function submit(e: SubmitEvent) {
										e.preventDefault()
										e.stopPropagation()
										if (!confirm('Are you sure you want to update this record?')) {
											return
										}

										const nr: api.RecordNew = {
											RelName: shortname(r.AbsName),
											Type: r.Type,
											TTL: parseInt(ttl.value),
											Value: value.value,
										}
										const xr = await check(fieldset, () => client.RecordUpdate(zone.Name, r.ID, nr))
										records.splice(records.indexOf(r), 1, xr)
										render()
										await refresh(fieldset)
									},
									fieldset=dom.fieldset(
										style({display: 'inline'}), // Prevent getting buttons on separate lines.
										dom.submitbutton('Update'),
									),
								),
								' ',
								dom.clickbutton('Delete', async function click(e: {target: HTMLButtonElement}) {
									if (!confirm('Are you sure you want to delete this record?')) {
										return
									}
									const xr = await check(e.target, () => client.RecordDelete(zone.Name, r.ID))
									records.splice(records.indexOf(r), 1, xr)
									render()
									await refresh(fieldset)
								}),
							),
					),
				)
			}),
		)
	}

	render()

	return root
}

const hashchange = async (e?: HashChangeEvent) => {
	const hash = decodeURIComponent(window.location.hash.substring(1))
	const t = hash.split('/')

	try {
		let elem: HTMLElement
		if (t.length === 1 && t[0] === '') {
			elem = await pageHome()
		} else if (t.length === 2 && t[0] === 'zones') {
			elem = await pageZone(t[1])
		} else {
			window.alert('Unknown hash')
			location.hash = '#'
			return
		}
		dom._kids(pageElem, elem)
	} catch (err: any) {
		window.alert('Error: '+err.message)
		window.location.hash = e?.oldURL ? new URL(e.oldURL).hash : ''
		throw err
	}
}

const init = async () => {
	[version, dnsTypeNames] = await Promise.all([
		client.Version(),
		client.DNSTypeNames(),
	])
	const root = dom.div(
		dom.div(
			style({display: 'flex', justifyContent: 'space-between', marginBottom: '1ex', padding: '.5em 1em', backgroundColor: '#f8f8f8'}),
			crumbElem,
			dom.div(
				dom.a(attr.href('https://github.com/mjl-/dnsclay'), 'dnsclay'),
				' ',
				version,
				' ',
				dom.a(attr.href('license'), 'license'),
			),
		),
		dom.div(
			pageElem,
		),
	)
	document.getElementById('rootElem')!.replaceWith(root)
	rootElem = root
	window.addEventListener('hashchange', hashchange)
	await hashchange()
}

window.addEventListener('load', async () => {
	try {
		await init()
	} catch (err: any) {
		window.alert('Error: ' + err.message)
	}
})
