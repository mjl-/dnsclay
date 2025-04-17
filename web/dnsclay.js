"use strict";
// Javascript is generated from typescript, do not modify generated javascript because changes will be overwritten.
const [dom, style, attr, prop] = (function () {
	// Start of unicode block (rough approximation of script), from https://www.unicode.org/Public/UNIDATA/Blocks.txt
	const scriptblocks = [0x0000, 0x0080, 0x0100, 0x0180, 0x0250, 0x02B0, 0x0300, 0x0370, 0x0400, 0x0500, 0x0530, 0x0590, 0x0600, 0x0700, 0x0750, 0x0780, 0x07C0, 0x0800, 0x0840, 0x0860, 0x0870, 0x08A0, 0x0900, 0x0980, 0x0A00, 0x0A80, 0x0B00, 0x0B80, 0x0C00, 0x0C80, 0x0D00, 0x0D80, 0x0E00, 0x0E80, 0x0F00, 0x1000, 0x10A0, 0x1100, 0x1200, 0x1380, 0x13A0, 0x1400, 0x1680, 0x16A0, 0x1700, 0x1720, 0x1740, 0x1760, 0x1780, 0x1800, 0x18B0, 0x1900, 0x1950, 0x1980, 0x19E0, 0x1A00, 0x1A20, 0x1AB0, 0x1B00, 0x1B80, 0x1BC0, 0x1C00, 0x1C50, 0x1C80, 0x1C90, 0x1CC0, 0x1CD0, 0x1D00, 0x1D80, 0x1DC0, 0x1E00, 0x1F00, 0x2000, 0x2070, 0x20A0, 0x20D0, 0x2100, 0x2150, 0x2190, 0x2200, 0x2300, 0x2400, 0x2440, 0x2460, 0x2500, 0x2580, 0x25A0, 0x2600, 0x2700, 0x27C0, 0x27F0, 0x2800, 0x2900, 0x2980, 0x2A00, 0x2B00, 0x2C00, 0x2C60, 0x2C80, 0x2D00, 0x2D30, 0x2D80, 0x2DE0, 0x2E00, 0x2E80, 0x2F00, 0x2FF0, 0x3000, 0x3040, 0x30A0, 0x3100, 0x3130, 0x3190, 0x31A0, 0x31C0, 0x31F0, 0x3200, 0x3300, 0x3400, 0x4DC0, 0x4E00, 0xA000, 0xA490, 0xA4D0, 0xA500, 0xA640, 0xA6A0, 0xA700, 0xA720, 0xA800, 0xA830, 0xA840, 0xA880, 0xA8E0, 0xA900, 0xA930, 0xA960, 0xA980, 0xA9E0, 0xAA00, 0xAA60, 0xAA80, 0xAAE0, 0xAB00, 0xAB30, 0xAB70, 0xABC0, 0xAC00, 0xD7B0, 0xD800, 0xDB80, 0xDC00, 0xE000, 0xF900, 0xFB00, 0xFB50, 0xFE00, 0xFE10, 0xFE20, 0xFE30, 0xFE50, 0xFE70, 0xFF00, 0xFFF0, 0x10000, 0x10080, 0x10100, 0x10140, 0x10190, 0x101D0, 0x10280, 0x102A0, 0x102E0, 0x10300, 0x10330, 0x10350, 0x10380, 0x103A0, 0x10400, 0x10450, 0x10480, 0x104B0, 0x10500, 0x10530, 0x10570, 0x10600, 0x10780, 0x10800, 0x10840, 0x10860, 0x10880, 0x108E0, 0x10900, 0x10920, 0x10980, 0x109A0, 0x10A00, 0x10A60, 0x10A80, 0x10AC0, 0x10B00, 0x10B40, 0x10B60, 0x10B80, 0x10C00, 0x10C80, 0x10D00, 0x10E60, 0x10E80, 0x10EC0, 0x10F00, 0x10F30, 0x10F70, 0x10FB0, 0x10FE0, 0x11000, 0x11080, 0x110D0, 0x11100, 0x11150, 0x11180, 0x111E0, 0x11200, 0x11280, 0x112B0, 0x11300, 0x11400, 0x11480, 0x11580, 0x11600, 0x11660, 0x11680, 0x11700, 0x11800, 0x118A0, 0x11900, 0x119A0, 0x11A00, 0x11A50, 0x11AB0, 0x11AC0, 0x11B00, 0x11C00, 0x11C70, 0x11D00, 0x11D60, 0x11EE0, 0x11F00, 0x11FB0, 0x11FC0, 0x12000, 0x12400, 0x12480, 0x12F90, 0x13000, 0x13430, 0x14400, 0x16800, 0x16A40, 0x16A70, 0x16AD0, 0x16B00, 0x16E40, 0x16F00, 0x16FE0, 0x17000, 0x18800, 0x18B00, 0x18D00, 0x1AFF0, 0x1B000, 0x1B100, 0x1B130, 0x1B170, 0x1BC00, 0x1BCA0, 0x1CF00, 0x1D000, 0x1D100, 0x1D200, 0x1D2C0, 0x1D2E0, 0x1D300, 0x1D360, 0x1D400, 0x1D800, 0x1DF00, 0x1E000, 0x1E030, 0x1E100, 0x1E290, 0x1E2C0, 0x1E4D0, 0x1E7E0, 0x1E800, 0x1E900, 0x1EC70, 0x1ED00, 0x1EE00, 0x1F000, 0x1F030, 0x1F0A0, 0x1F100, 0x1F200, 0x1F300, 0x1F600, 0x1F650, 0x1F680, 0x1F700, 0x1F780, 0x1F800, 0x1F900, 0x1FA00, 0x1FA70, 0x1FB00, 0x20000, 0x2A700, 0x2B740, 0x2B820, 0x2CEB0, 0x2F800, 0x30000, 0x31350, 0xE0000, 0xE0100, 0xF0000, 0x100000];
	// Find block code belongs in.
	const findBlock = (code) => {
		let s = 0;
		let e = scriptblocks.length;
		while (s < e - 1) {
			let i = Math.floor((s + e) / 2);
			if (code < scriptblocks[i]) {
				e = i;
			}
			else {
				s = i;
			}
		}
		return s;
	};
	// formatText adds s to element e, in a way that makes switching unicode scripts
	// clear, with alternating DOM TextNode and span elements with a "switchscript"
	// class. Useful for highlighting look alikes, e.g. a (ascii 0x61) and а (cyrillic
	// 0x430).
	//
	// This is only called one string at a time, so the UI can still display strings
	// without highlighting switching scripts, by calling formatText on the parts.
	const formatText = (e, s) => {
		// Handle some common cases quickly.
		if (!s) {
			return;
		}
		let ascii = true;
		for (const c of s) {
			const cp = c.codePointAt(0); // For typescript, to check for undefined.
			if (cp !== undefined && cp >= 0x0080) {
				ascii = false;
				break;
			}
		}
		if (ascii) {
			e.appendChild(document.createTextNode(s));
			return;
		}
		// todo: handle grapheme clusters? wait for Intl.Segmenter?
		let n = 0; // Number of text/span parts added.
		let str = ''; // Collected so far.
		let block = -1; // Previous block/script.
		let mod = 1;
		const put = (nextblock) => {
			if (n === 0 && nextblock === 0) {
				// Start was non-ascii, second block is ascii, we'll start marked as switched.
				mod = 0;
			}
			if (n % 2 === mod) {
				const x = document.createElement('span');
				x.classList.add('scriptswitch');
				x.appendChild(document.createTextNode(str));
				e.appendChild(x);
			}
			else {
				e.appendChild(document.createTextNode(str));
			}
			n++;
			str = '';
		};
		for (const c of s) {
			// Basic whitespace does not switch blocks. Will probably need to extend with more
			// punctuation in the future. Possibly for digits too. But perhaps not in all
			// scripts.
			if (c === ' ' || c === '\t' || c === '\r' || c === '\n') {
				str += c;
				continue;
			}
			const code = c.codePointAt(0);
			if (block < 0 || !(code >= scriptblocks[block] && (code < scriptblocks[block + 1] || block === scriptblocks.length - 1))) {
				const nextblock = code < 0x0080 ? 0 : findBlock(code);
				if (block >= 0) {
					put(nextblock);
				}
				block = nextblock;
			}
			str += c;
		}
		put(-1);
	};
	const _domKids = (e, l) => {
		l.forEach((c) => {
			const xc = c;
			if (typeof c === 'string') {
				formatText(e, c);
			}
			else if (c instanceof String) {
				// String is an escape-hatch for text that should not be formatted with
				// unicode-block-change-highlighting, e.g. for textarea values.
				e.appendChild(document.createTextNode('' + c));
			}
			else if (c instanceof Element) {
				e.appendChild(c);
			}
			else if (c instanceof Function) {
				if (!c.name) {
					throw new Error('function without name');
				}
				e.addEventListener(c.name, c);
			}
			else if (Array.isArray(xc)) {
				_domKids(e, c);
			}
			else if (xc._class) {
				for (const s of xc._class) {
					e.classList.toggle(s, true);
				}
			}
			else if (xc._attrs) {
				for (const k in xc._attrs) {
					e.setAttribute(k, xc._attrs[k]);
				}
			}
			else if (xc._styles) {
				for (const k in xc._styles) {
					const estyle = e.style;
					estyle[k] = xc._styles[k];
				}
			}
			else if (xc._props) {
				for (const k in xc._props) {
					const eprops = e;
					eprops[k] = xc._props[k];
				}
			}
			else if (xc.root) {
				e.appendChild(xc.root);
			}
			else {
				console.log('bad kid', c);
				throw new Error('bad kid');
			}
		});
		return e;
	};
	const dom = {
		_kids: function (e, ...kl) {
			while (e.firstChild) {
				e.removeChild(e.firstChild);
			}
			_domKids(e, kl);
		},
		_attrs: (x) => { return { _attrs: x }; },
		_class: (...x) => { return { _class: x }; },
		// The createElement calls are spelled out so typescript can derive function
		// signatures with a specific HTML*Element return type.
		div: (...l) => _domKids(document.createElement('div'), l),
		span: (...l) => _domKids(document.createElement('span'), l),
		a: (...l) => _domKids(document.createElement('a'), l),
		input: (...l) => _domKids(document.createElement('input'), l),
		textarea: (...l) => _domKids(document.createElement('textarea'), l),
		select: (...l) => _domKids(document.createElement('select'), l),
		option: (...l) => _domKids(document.createElement('option'), l),
		clickbutton: (...l) => _domKids(document.createElement('button'), [attr.type('button'), ...l]),
		submitbutton: (...l) => _domKids(document.createElement('button'), [attr.type('submit'), ...l]),
		form: (...l) => _domKids(document.createElement('form'), l),
		fieldset: (...l) => _domKids(document.createElement('fieldset'), l),
		table: (...l) => _domKids(document.createElement('table'), l),
		thead: (...l) => _domKids(document.createElement('thead'), l),
		tbody: (...l) => _domKids(document.createElement('tbody'), l),
		tfoot: (...l) => _domKids(document.createElement('tfoot'), l),
		tr: (...l) => _domKids(document.createElement('tr'), l),
		td: (...l) => _domKids(document.createElement('td'), l),
		th: (...l) => _domKids(document.createElement('th'), l),
		datalist: (...l) => _domKids(document.createElement('datalist'), l),
		h1: (...l) => _domKids(document.createElement('h1'), l),
		h2: (...l) => _domKids(document.createElement('h2'), l),
		h3: (...l) => _domKids(document.createElement('h3'), l),
		br: (...l) => _domKids(document.createElement('br'), l),
		hr: (...l) => _domKids(document.createElement('hr'), l),
		pre: (...l) => _domKids(document.createElement('pre'), l),
		label: (...l) => _domKids(document.createElement('label'), l),
		ul: (...l) => _domKids(document.createElement('ul'), l),
		li: (...l) => _domKids(document.createElement('li'), l),
		iframe: (...l) => _domKids(document.createElement('iframe'), l),
		b: (...l) => _domKids(document.createElement('b'), l),
		img: (...l) => _domKids(document.createElement('img'), l),
		style: (...l) => _domKids(document.createElement('style'), l),
		search: (...l) => _domKids(document.createElement('search'), l),
		p: (...l) => _domKids(document.createElement('p'), l),
		tt: (...l) => _domKids(document.createElement('tt'), l),
		i: (...l) => _domKids(document.createElement('i'), l),
		link: (...l) => _domKids(document.createElement('link'), l),
		optgroup: (...l) => _domKids(document.createElement('optgroup'), l),
	};
	const _attr = (k, v) => { const o = {}; o[k] = v; return { _attrs: o }; };
	const attr = {
		title: (s) => _attr('title', s),
		value: (s) => _attr('value', s),
		type: (s) => _attr('type', s),
		tabindex: (s) => _attr('tabindex', s),
		src: (s) => _attr('src', s),
		placeholder: (s) => _attr('placeholder', s),
		href: (s) => _attr('href', s),
		checked: (s) => _attr('checked', s),
		selected: (s) => _attr('selected', s),
		id: (s) => _attr('id', s),
		datalist: (s) => _attr('datalist', s),
		rows: (s) => _attr('rows', s),
		target: (s) => _attr('target', s),
		rel: (s) => _attr('rel', s),
		required: (s) => _attr('required', s),
		multiple: (s) => _attr('multiple', s),
		download: (s) => _attr('download', s),
		disabled: (s) => _attr('disabled', s),
		draggable: (s) => _attr('draggable', s),
		rowspan: (s) => _attr('rowspan', s),
		colspan: (s) => _attr('colspan', s),
		for: (s) => _attr('for', s),
		role: (s) => _attr('role', s),
		arialabel: (s) => _attr('aria-label', s),
		arialive: (s) => _attr('aria-live', s),
		name: (s) => _attr('name', s),
		min: (s) => _attr('min', s),
		max: (s) => _attr('max', s),
		action: (s) => _attr('action', s),
		method: (s) => _attr('method', s),
		autocomplete: (s) => _attr('autocomplete', s),
		list: (s) => _attr('list', s),
		form: (s) => _attr('form', s),
		size: (s) => _attr('size', s),
		label: (s) => _attr('label', s),
	};
	const style = (x) => { return { _styles: x }; };
	const prop = (x) => { return { _props: x }; };
	return [dom, style, attr, prop];
})();
// NOTE: GENERATED by github.com/mjl-/sherpats, DO NOT MODIFY
var api;
(function (api) {
	let BaseURL;
	(function (BaseURL) {
		BaseURL["Sandbox"] = "https://api.sandbox.dnsmadeeasy.com/V2.0/";
		BaseURL["Prod"] = "https://api.dnsmadeeasy.com/V2.0/";
	})(BaseURL = api.BaseURL || (api.BaseURL = {}));
	api.structTypes = { "AuthOpenStack": true, "Credential": true, "IntValue": true, "KnownProviders": true, "PropagationState": true, "Provider": true, "ProviderConfig": true, "Provider_alidns": true, "Provider_autodns": true, "Provider_azure": true, "Provider_bunny": true, "Provider_civo": true, "Provider_cloudflare": true, "Provider_cloudns": true, "Provider_ddnss": true, "Provider_desec": true, "Provider_digitalocean": true, "Provider_directadmin": true, "Provider_dnsimple": true, "Provider_dnsmadeeasy": true, "Provider_dnspod": true, "Provider_dnsupdate": true, "Provider_domainnameshop": true, "Provider_dreamhost": true, "Provider_duckdns": true, "Provider_dynu": true, "Provider_dynv6": true, "Provider_easydns": true, "Provider_exoscale": true, "Provider_gandi": true, "Provider_gcore": true, "Provider_glesys": true, "Provider_godaddy": true, "Provider_googleclouddns": true, "Provider_he": true, "Provider_hetzner": true, "Provider_hexonet": true, "Provider_hosttech": true, "Provider_huaweicloud": true, "Provider_infomaniak": true, "Provider_inwx": true, "Provider_ionos": true, "Provider_katapult": true, "Provider_leaseweb": true, "Provider_linode": true, "Provider_loopia": true, "Provider_luadns": true, "Provider_mailinabox": true, "Provider_metaname": true, "Provider_mythicbeasts": true, "Provider_namecheap": true, "Provider_namedotcom": true, "Provider_namesilo": true, "Provider_nanelo": true, "Provider_netcup": true, "Provider_netlify": true, "Provider_nfsn": true, "Provider_njalla": true, "Provider_ovh": true, "Provider_porkbun": true, "Provider_powerdns": true, "Provider_rfc2136": true, "Provider_route53": true, "Provider_scaleway": true, "Provider_selectel": true, "Provider_tencentcloud": true, "Provider_timeweb": true, "Provider_totaluptime": true, "Provider_vultr": true, "Record": true, "RecordSet": true, "RecordSetChange": true, "StringValue": true, "Zone": true, "ZoneNotify": true, "sherpadocArg": true, "sherpadocField": true, "sherpadocFunction": true, "sherpadocInts": true, "sherpadocSection": true, "sherpadocStrings": true, "sherpadocStruct": true };
	api.stringsTypes = { "BaseURL": true };
	api.intsTypes = {};
	api.types = {
		"Zone": { "Name": "Zone", "Docs": "", "Fields": [{ "Name": "Name", "Docs": "", "Typewords": ["string"] }, { "Name": "ProviderConfigName", "Docs": "", "Typewords": ["string"] }, { "Name": "SerialLocal", "Docs": "", "Typewords": ["uint32"] }, { "Name": "SerialRemote", "Docs": "", "Typewords": ["uint32"] }, { "Name": "LastSync", "Docs": "", "Typewords": ["nullable", "timestamp"] }, { "Name": "LastRecordChange", "Docs": "", "Typewords": ["nullable", "timestamp"] }, { "Name": "SyncInterval", "Docs": "", "Typewords": ["int64"] }, { "Name": "RefreshInterval", "Docs": "", "Typewords": ["int64"] }, { "Name": "NextSync", "Docs": "", "Typewords": ["timestamp"] }, { "Name": "NextRefresh", "Docs": "", "Typewords": ["timestamp"] }] },
		"ProviderConfig": { "Name": "ProviderConfig", "Docs": "", "Fields": [{ "Name": "Name", "Docs": "", "Typewords": ["string"] }, { "Name": "ProviderName", "Docs": "", "Typewords": ["string"] }, { "Name": "ProviderConfigJSON", "Docs": "", "Typewords": ["string"] }] },
		"ZoneNotify": { "Name": "ZoneNotify", "Docs": "", "Fields": [{ "Name": "ID", "Docs": "", "Typewords": ["int64"] }, { "Name": "Created", "Docs": "", "Typewords": ["timestamp"] }, { "Name": "Zone", "Docs": "", "Typewords": ["string"] }, { "Name": "Address", "Docs": "", "Typewords": ["string"] }, { "Name": "Protocol", "Docs": "", "Typewords": ["string"] }] },
		"Credential": { "Name": "Credential", "Docs": "", "Fields": [{ "Name": "ID", "Docs": "", "Typewords": ["int64"] }, { "Name": "Created", "Docs": "", "Typewords": ["timestamp"] }, { "Name": "Name", "Docs": "", "Typewords": ["string"] }, { "Name": "Type", "Docs": "", "Typewords": ["string"] }, { "Name": "TSIGSecret", "Docs": "", "Typewords": ["string"] }, { "Name": "TLSPublicKey", "Docs": "", "Typewords": ["string"] }] },
		"RecordSet": { "Name": "RecordSet", "Docs": "", "Fields": [{ "Name": "Records", "Docs": "", "Typewords": ["[]", "Record"] }, { "Name": "States", "Docs": "", "Typewords": ["[]", "PropagationState"] }] },
		"Record": { "Name": "Record", "Docs": "", "Fields": [{ "Name": "ID", "Docs": "", "Typewords": ["int64"] }, { "Name": "Zone", "Docs": "", "Typewords": ["string"] }, { "Name": "SerialFirst", "Docs": "", "Typewords": ["uint32"] }, { "Name": "SerialDeleted", "Docs": "", "Typewords": ["uint32"] }, { "Name": "First", "Docs": "", "Typewords": ["timestamp"] }, { "Name": "Deleted", "Docs": "", "Typewords": ["nullable", "timestamp"] }, { "Name": "AbsName", "Docs": "", "Typewords": ["string"] }, { "Name": "Type", "Docs": "", "Typewords": ["uint16"] }, { "Name": "Class", "Docs": "", "Typewords": ["uint16"] }, { "Name": "TTL", "Docs": "", "Typewords": ["uint32"] }, { "Name": "DataHex", "Docs": "", "Typewords": ["string"] }, { "Name": "Value", "Docs": "", "Typewords": ["string"] }, { "Name": "ProviderID", "Docs": "", "Typewords": ["string"] }] },
		"PropagationState": { "Name": "PropagationState", "Docs": "", "Fields": [{ "Name": "Start", "Docs": "", "Typewords": ["timestamp"] }, { "Name": "End", "Docs": "", "Typewords": ["nullable", "timestamp"] }, { "Name": "Negative", "Docs": "", "Typewords": ["bool"] }, { "Name": "Records", "Docs": "", "Typewords": ["[]", "Record"] }] },
		"RecordSetChange": { "Name": "RecordSetChange", "Docs": "", "Fields": [{ "Name": "RelName", "Docs": "", "Typewords": ["string"] }, { "Name": "TTL", "Docs": "", "Typewords": ["uint32"] }, { "Name": "Type", "Docs": "", "Typewords": ["uint16"] }, { "Name": "Values", "Docs": "", "Typewords": ["[]", "string"] }] },
		"KnownProviders": { "Name": "KnownProviders", "Docs": "", "Fields": [{ "Name": "Xalidns", "Docs": "", "Typewords": ["Provider_alidns"] }, { "Name": "Xautodns", "Docs": "", "Typewords": ["Provider_autodns"] }, { "Name": "Xazure", "Docs": "", "Typewords": ["Provider_azure"] }, { "Name": "Xbunny", "Docs": "", "Typewords": ["Provider_bunny"] }, { "Name": "Xcivo", "Docs": "", "Typewords": ["Provider_civo"] }, { "Name": "Xcloudflare", "Docs": "", "Typewords": ["Provider_cloudflare"] }, { "Name": "Xcloudns", "Docs": "", "Typewords": ["Provider_cloudns"] }, { "Name": "Xddnss", "Docs": "", "Typewords": ["Provider_ddnss"] }, { "Name": "Xdesec", "Docs": "", "Typewords": ["Provider_desec"] }, { "Name": "Xdigitalocean", "Docs": "", "Typewords": ["Provider_digitalocean"] }, { "Name": "Xdirectadmin", "Docs": "", "Typewords": ["Provider_directadmin"] }, { "Name": "Xdnsimple", "Docs": "", "Typewords": ["Provider_dnsimple"] }, { "Name": "Xdnsmadeeasy", "Docs": "", "Typewords": ["Provider_dnsmadeeasy"] }, { "Name": "Xdnspod", "Docs": "", "Typewords": ["Provider_dnspod"] }, { "Name": "Xdnsupdate", "Docs": "", "Typewords": ["Provider_dnsupdate"] }, { "Name": "Xdomainnameshop", "Docs": "", "Typewords": ["Provider_domainnameshop"] }, { "Name": "Xdreamhost", "Docs": "", "Typewords": ["Provider_dreamhost"] }, { "Name": "Xduckdns", "Docs": "", "Typewords": ["Provider_duckdns"] }, { "Name": "Xdynu", "Docs": "", "Typewords": ["Provider_dynu"] }, { "Name": "Xdynv6", "Docs": "", "Typewords": ["Provider_dynv6"] }, { "Name": "Xeasydns", "Docs": "", "Typewords": ["Provider_easydns"] }, { "Name": "Xexoscale", "Docs": "", "Typewords": ["Provider_exoscale"] }, { "Name": "Xgandi", "Docs": "", "Typewords": ["Provider_gandi"] }, { "Name": "Xgcore", "Docs": "", "Typewords": ["Provider_gcore"] }, { "Name": "Xglesys", "Docs": "", "Typewords": ["Provider_glesys"] }, { "Name": "Xgodaddy", "Docs": "", "Typewords": ["Provider_godaddy"] }, { "Name": "Xgoogleclouddns", "Docs": "", "Typewords": ["Provider_googleclouddns"] }, { "Name": "Xhe", "Docs": "", "Typewords": ["Provider_he"] }, { "Name": "Xhetzner", "Docs": "", "Typewords": ["Provider_hetzner"] }, { "Name": "Xhexonet", "Docs": "", "Typewords": ["Provider_hexonet"] }, { "Name": "Xhosttech", "Docs": "", "Typewords": ["Provider_hosttech"] }, { "Name": "Xhuaweicloud", "Docs": "", "Typewords": ["Provider_huaweicloud"] }, { "Name": "Xinfomaniak", "Docs": "", "Typewords": ["Provider_infomaniak"] }, { "Name": "Xinwx", "Docs": "", "Typewords": ["Provider_inwx"] }, { "Name": "Xionos", "Docs": "", "Typewords": ["Provider_ionos"] }, { "Name": "Xkatapult", "Docs": "", "Typewords": ["Provider_katapult"] }, { "Name": "Xleaseweb", "Docs": "", "Typewords": ["Provider_leaseweb"] }, { "Name": "Xlinode", "Docs": "", "Typewords": ["Provider_linode"] }, { "Name": "Xloopia", "Docs": "", "Typewords": ["Provider_loopia"] }, { "Name": "Xluadns", "Docs": "", "Typewords": ["Provider_luadns"] }, { "Name": "Xmailinabox", "Docs": "", "Typewords": ["Provider_mailinabox"] }, { "Name": "Xmetaname", "Docs": "", "Typewords": ["Provider_metaname"] }, { "Name": "Xmythicbeasts", "Docs": "", "Typewords": ["Provider_mythicbeasts"] }, { "Name": "Xnamecheap", "Docs": "", "Typewords": ["Provider_namecheap"] }, { "Name": "Xnamedotcom", "Docs": "", "Typewords": ["Provider_namedotcom"] }, { "Name": "Xnamesilo", "Docs": "", "Typewords": ["Provider_namesilo"] }, { "Name": "Xnanelo", "Docs": "", "Typewords": ["Provider_nanelo"] }, { "Name": "Xnetcup", "Docs": "", "Typewords": ["Provider_netcup"] }, { "Name": "Xnetlify", "Docs": "", "Typewords": ["Provider_netlify"] }, { "Name": "Xnfsn", "Docs": "", "Typewords": ["Provider_nfsn"] }, { "Name": "Xnjalla", "Docs": "", "Typewords": ["Provider_njalla"] }, { "Name": "Xopenstackdesignate", "Docs": "", "Typewords": ["Provider"] }, { "Name": "Xovh", "Docs": "", "Typewords": ["Provider_ovh"] }, { "Name": "Xporkbun", "Docs": "", "Typewords": ["Provider_porkbun"] }, { "Name": "Xpowerdns", "Docs": "", "Typewords": ["Provider_powerdns"] }, { "Name": "Xrfc2136", "Docs": "", "Typewords": ["Provider_rfc2136"] }, { "Name": "Xroute53", "Docs": "", "Typewords": ["Provider_route53"] }, { "Name": "Xscaleway", "Docs": "", "Typewords": ["Provider_scaleway"] }, { "Name": "Xselectel", "Docs": "", "Typewords": ["Provider_selectel"] }, { "Name": "Xtencentcloud", "Docs": "", "Typewords": ["Provider_tencentcloud"] }, { "Name": "Xtimeweb", "Docs": "", "Typewords": ["Provider_timeweb"] }, { "Name": "Xtotaluptime", "Docs": "", "Typewords": ["Provider_totaluptime"] }, { "Name": "Xvultr", "Docs": "", "Typewords": ["Provider_vultr"] }] },
		"Provider_alidns": { "Name": "Provider_alidns", "Docs": "", "Fields": [{ "Name": "access_key_id", "Docs": "", "Typewords": ["string"] }, { "Name": "access_key_secret", "Docs": "", "Typewords": ["string"] }, { "Name": "region_id", "Docs": "", "Typewords": ["nullable", "string"] }] },
		"Provider_autodns": { "Name": "Provider_autodns", "Docs": "", "Fields": [{ "Name": "username", "Docs": "", "Typewords": ["string"] }, { "Name": "password", "Docs": "", "Typewords": ["string"] }, { "Name": "Endpoint", "Docs": "", "Typewords": ["string"] }, { "Name": "context", "Docs": "", "Typewords": ["string"] }, { "Name": "primary", "Docs": "", "Typewords": ["string"] }] },
		"Provider_azure": { "Name": "Provider_azure", "Docs": "", "Fields": [{ "Name": "subscription_id", "Docs": "", "Typewords": ["nullable", "string"] }, { "Name": "resource_group_name", "Docs": "", "Typewords": ["nullable", "string"] }, { "Name": "tenant_id", "Docs": "", "Typewords": ["nullable", "string"] }, { "Name": "client_id", "Docs": "", "Typewords": ["nullable", "string"] }, { "Name": "client_secret", "Docs": "", "Typewords": ["nullable", "string"] }] },
		"Provider_bunny": { "Name": "Provider_bunny", "Docs": "", "Fields": [{ "Name": "access_key", "Docs": "", "Typewords": ["string"] }, { "Name": "debug", "Docs": "", "Typewords": ["bool"] }] },
		"Provider_civo": { "Name": "Provider_civo", "Docs": "", "Fields": [{ "Name": "api_token", "Docs": "", "Typewords": ["nullable", "string"] }] },
		"Provider_cloudflare": { "Name": "Provider_cloudflare", "Docs": "", "Fields": [{ "Name": "api_token", "Docs": "", "Typewords": ["nullable", "string"] }, { "Name": "zone_token", "Docs": "", "Typewords": ["nullable", "string"] }] },
		"Provider_cloudns": { "Name": "Provider_cloudns", "Docs": "", "Fields": [{ "Name": "auth_id", "Docs": "", "Typewords": ["string"] }, { "Name": "sub_auth_id", "Docs": "", "Typewords": ["nullable", "string"] }, { "Name": "auth_password", "Docs": "", "Typewords": ["string"] }] },
		"Provider_ddnss": { "Name": "Provider_ddnss", "Docs": "", "Fields": [{ "Name": "api_token", "Docs": "", "Typewords": ["string"] }, { "Name": "username", "Docs": "", "Typewords": ["nullable", "string"] }, { "Name": "password", "Docs": "", "Typewords": ["nullable", "string"] }] },
		"Provider_desec": { "Name": "Provider_desec", "Docs": "", "Fields": [{ "Name": "token", "Docs": "", "Typewords": ["nullable", "string"] }] },
		"Provider_digitalocean": { "Name": "Provider_digitalocean", "Docs": "", "Fields": [{ "Name": "auth_token", "Docs": "", "Typewords": ["string"] }] },
		"Provider_directadmin": { "Name": "Provider_directadmin", "Docs": "", "Fields": [{ "Name": "host", "Docs": "", "Typewords": ["nullable", "string"] }, { "Name": "user", "Docs": "", "Typewords": ["nullable", "string"] }, { "Name": "login_key", "Docs": "", "Typewords": ["nullable", "string"] }, { "Name": "insecure_requests", "Docs": "", "Typewords": ["nullable", "bool"] }, { "Name": "debug", "Docs": "", "Typewords": ["nullable", "string"] }] },
		"Provider_dnsimple": { "Name": "Provider_dnsimple", "Docs": "", "Fields": [{ "Name": "api_access_token", "Docs": "", "Typewords": ["nullable", "string"] }, { "Name": "account_id", "Docs": "", "Typewords": ["nullable", "string"] }, { "Name": "api_url", "Docs": "", "Typewords": ["nullable", "string"] }] },
		"Provider_dnsmadeeasy": { "Name": "Provider_dnsmadeeasy", "Docs": "", "Fields": [{ "Name": "api_key", "Docs": "", "Typewords": ["nullable", "string"] }, { "Name": "secret_key", "Docs": "", "Typewords": ["nullable", "string"] }, { "Name": "api_endpoint", "Docs": "", "Typewords": ["nullable", "BaseURL"] }] },
		"Provider_dnspod": { "Name": "Provider_dnspod", "Docs": "", "Fields": [{ "Name": "auth_token", "Docs": "", "Typewords": ["string"] }] },
		"Provider_dnsupdate": { "Name": "Provider_dnsupdate", "Docs": "", "Fields": [{ "Name": "addr", "Docs": "", "Typewords": ["nullable", "string"] }] },
		"Provider_domainnameshop": { "Name": "Provider_domainnameshop", "Docs": "", "Fields": [{ "Name": "api_token", "Docs": "", "Typewords": ["string"] }, { "Name": "api_secret", "Docs": "", "Typewords": ["string"] }] },
		"Provider_dreamhost": { "Name": "Provider_dreamhost", "Docs": "", "Fields": [{ "Name": "api_key", "Docs": "", "Typewords": ["nullable", "string"] }] },
		"Provider_duckdns": { "Name": "Provider_duckdns", "Docs": "", "Fields": [{ "Name": "api_token", "Docs": "", "Typewords": ["nullable", "string"] }, { "Name": "override_domain", "Docs": "", "Typewords": ["nullable", "string"] }] },
		"Provider_dynu": { "Name": "Provider_dynu", "Docs": "", "Fields": [{ "Name": "api_token", "Docs": "", "Typewords": ["nullable", "string"] }, { "Name": "own_domain", "Docs": "", "Typewords": ["nullable", "string"] }] },
		"Provider_dynv6": { "Name": "Provider_dynv6", "Docs": "", "Fields": [{ "Name": "token", "Docs": "", "Typewords": ["nullable", "string"] }] },
		"Provider_easydns": { "Name": "Provider_easydns", "Docs": "", "Fields": [{ "Name": "api_token", "Docs": "", "Typewords": ["nullable", "string"] }, { "Name": "api_key", "Docs": "", "Typewords": ["nullable", "string"] }, { "Name": "api_url", "Docs": "", "Typewords": ["nullable", "string"] }] },
		"Provider_exoscale": { "Name": "Provider_exoscale", "Docs": "", "Fields": [{ "Name": "api_key", "Docs": "", "Typewords": ["nullable", "string"] }, { "Name": "api_secret", "Docs": "", "Typewords": ["nullable", "string"] }] },
		"Provider_gandi": { "Name": "Provider_gandi", "Docs": "", "Fields": [{ "Name": "bearer_token", "Docs": "", "Typewords": ["nullable", "string"] }] },
		"Provider_gcore": { "Name": "Provider_gcore", "Docs": "", "Fields": [{ "Name": "api_key", "Docs": "", "Typewords": ["nullable", "string"] }] },
		"Provider_glesys": { "Name": "Provider_glesys", "Docs": "", "Fields": [{ "Name": "project", "Docs": "", "Typewords": ["nullable", "string"] }, { "Name": "api_key", "Docs": "", "Typewords": ["nullable", "string"] }] },
		"Provider_godaddy": { "Name": "Provider_godaddy", "Docs": "", "Fields": [{ "Name": "api_token", "Docs": "", "Typewords": ["nullable", "string"] }] },
		"Provider_googleclouddns": { "Name": "Provider_googleclouddns", "Docs": "", "Fields": [{ "Name": "gcp_project", "Docs": "", "Typewords": ["nullable", "string"] }, { "Name": "gcp_application_default", "Docs": "", "Typewords": ["nullable", "string"] }] },
		"Provider_he": { "Name": "Provider_he", "Docs": "", "Fields": [{ "Name": "api_key", "Docs": "", "Typewords": ["nullable", "string"] }] },
		"Provider_hetzner": { "Name": "Provider_hetzner", "Docs": "", "Fields": [{ "Name": "auth_api_token", "Docs": "", "Typewords": ["string"] }] },
		"Provider_hexonet": { "Name": "Provider_hexonet", "Docs": "", "Fields": [{ "Name": "username", "Docs": "", "Typewords": ["string"] }, { "Name": "password", "Docs": "", "Typewords": ["nullable", "string"] }, { "Name": "debug", "Docs": "", "Typewords": ["nullable", "string"] }] },
		"Provider_hosttech": { "Name": "Provider_hosttech", "Docs": "", "Fields": [{ "Name": "api_token", "Docs": "", "Typewords": ["nullable", "string"] }] },
		"Provider_huaweicloud": { "Name": "Provider_huaweicloud", "Docs": "", "Fields": [{ "Name": "access_key_id", "Docs": "", "Typewords": ["nullable", "string"] }, { "Name": "secret_access_key", "Docs": "", "Typewords": ["nullable", "string"] }, { "Name": "region_id", "Docs": "", "Typewords": ["nullable", "string"] }] },
		"Provider_infomaniak": { "Name": "Provider_infomaniak", "Docs": "", "Fields": [{ "Name": "api_token", "Docs": "", "Typewords": ["nullable", "string"] }] },
		"Provider_inwx": { "Name": "Provider_inwx", "Docs": "", "Fields": [{ "Name": "username", "Docs": "", "Typewords": ["nullable", "string"] }, { "Name": "password", "Docs": "", "Typewords": ["nullable", "string"] }, { "Name": "shared_secret", "Docs": "", "Typewords": ["nullable", "string"] }, { "Name": "endpoint_url", "Docs": "", "Typewords": ["nullable", "string"] }] },
		"Provider_ionos": { "Name": "Provider_ionos", "Docs": "", "Fields": [{ "Name": "auth_api_token", "Docs": "", "Typewords": ["string"] }] },
		"Provider_katapult": { "Name": "Provider_katapult", "Docs": "", "Fields": [{ "Name": "api_token", "Docs": "", "Typewords": ["nullable", "string"] }] },
		"Provider_leaseweb": { "Name": "Provider_leaseweb", "Docs": "", "Fields": [{ "Name": "api_token", "Docs": "", "Typewords": ["nullable", "string"] }] },
		"Provider_linode": { "Name": "Provider_linode", "Docs": "", "Fields": [{ "Name": "api_token", "Docs": "", "Typewords": ["nullable", "string"] }, { "Name": "api_url", "Docs": "", "Typewords": ["nullable", "string"] }, { "Name": "api_version", "Docs": "", "Typewords": ["nullable", "string"] }] },
		"Provider_loopia": { "Name": "Provider_loopia", "Docs": "", "Fields": [{ "Name": "username", "Docs": "", "Typewords": ["nullable", "string"] }, { "Name": "password", "Docs": "", "Typewords": ["nullable", "string"] }, { "Name": "customer", "Docs": "", "Typewords": ["nullable", "string"] }] },
		"Provider_luadns": { "Name": "Provider_luadns", "Docs": "", "Fields": [{ "Name": "email", "Docs": "", "Typewords": ["nullable", "string"] }, { "Name": "api_key", "Docs": "", "Typewords": ["nullable", "string"] }] },
		"Provider_mailinabox": { "Name": "Provider_mailinabox", "Docs": "", "Fields": [{ "Name": "api_url", "Docs": "", "Typewords": ["nullable", "string"] }, { "Name": "email_address", "Docs": "", "Typewords": ["nullable", "string"] }, { "Name": "password", "Docs": "", "Typewords": ["nullable", "string"] }] },
		"Provider_metaname": { "Name": "Provider_metaname", "Docs": "", "Fields": [{ "Name": "api_key", "Docs": "", "Typewords": ["nullable", "string"] }, { "Name": "account_reference", "Docs": "", "Typewords": ["nullable", "string"] }, { "Name": "endpoint", "Docs": "", "Typewords": ["nullable", "string"] }] },
		"Provider_mythicbeasts": { "Name": "Provider_mythicbeasts", "Docs": "", "Fields": [{ "Name": "key_id", "Docs": "", "Typewords": ["nullable", "string"] }, { "Name": "secret", "Docs": "", "Typewords": ["nullable", "string"] }] },
		"Provider_namecheap": { "Name": "Provider_namecheap", "Docs": "", "Fields": [{ "Name": "api_key", "Docs": "", "Typewords": ["nullable", "string"] }, { "Name": "user", "Docs": "", "Typewords": ["nullable", "string"] }, { "Name": "api_endpoint", "Docs": "", "Typewords": ["nullable", "string"] }, { "Name": "client_ip", "Docs": "", "Typewords": ["nullable", "string"] }] },
		"Provider_namedotcom": { "Name": "Provider_namedotcom", "Docs": "", "Fields": [{ "Name": "api_token", "Docs": "", "Typewords": ["nullable", "string"] }, { "Name": "user", "Docs": "", "Typewords": ["nullable", "string"] }, { "Name": "server", "Docs": "", "Typewords": ["nullable", "string"] }] },
		"Provider_namesilo": { "Name": "Provider_namesilo", "Docs": "", "Fields": [{ "Name": "api_token", "Docs": "", "Typewords": ["nullable", "string"] }] },
		"Provider_nanelo": { "Name": "Provider_nanelo", "Docs": "", "Fields": [{ "Name": "api_token", "Docs": "", "Typewords": ["nullable", "string"] }] },
		"Provider_netcup": { "Name": "Provider_netcup", "Docs": "", "Fields": [{ "Name": "customer_number", "Docs": "", "Typewords": ["string"] }, { "Name": "api_key", "Docs": "", "Typewords": ["string"] }, { "Name": "api_password", "Docs": "", "Typewords": ["string"] }] },
		"Provider_netlify": { "Name": "Provider_netlify", "Docs": "", "Fields": [{ "Name": "personal_access_token", "Docs": "", "Typewords": ["nullable", "string"] }] },
		"Provider_nfsn": { "Name": "Provider_nfsn", "Docs": "", "Fields": [{ "Name": "login", "Docs": "", "Typewords": ["nullable", "string"] }, { "Name": "api_key", "Docs": "", "Typewords": ["nullable", "string"] }] },
		"Provider_njalla": { "Name": "Provider_njalla", "Docs": "", "Fields": [{ "Name": "api_token", "Docs": "", "Typewords": ["nullable", "string"] }] },
		"Provider": { "Name": "Provider", "Docs": "", "Fields": [{ "Name": "auth_open_stack", "Docs": "", "Typewords": ["AuthOpenStack"] }] },
		"AuthOpenStack": { "Name": "AuthOpenStack", "Docs": "", "Fields": [{ "Name": "region_name", "Docs": "", "Typewords": ["string"] }, { "Name": "tenant_id", "Docs": "", "Typewords": ["string"] }, { "Name": "identity_api_version", "Docs": "", "Typewords": ["string"] }, { "Name": "password", "Docs": "", "Typewords": ["string"] }, { "Name": "auth_url", "Docs": "", "Typewords": ["string"] }, { "Name": "username", "Docs": "", "Typewords": ["string"] }, { "Name": "tenant_name", "Docs": "", "Typewords": ["string"] }, { "Name": "endpoint_type", "Docs": "", "Typewords": ["string"] }] },
		"Provider_ovh": { "Name": "Provider_ovh", "Docs": "", "Fields": [{ "Name": "endpoint", "Docs": "", "Typewords": ["nullable", "string"] }, { "Name": "application_key", "Docs": "", "Typewords": ["nullable", "string"] }, { "Name": "application_secret", "Docs": "", "Typewords": ["nullable", "string"] }, { "Name": "consumer_key", "Docs": "", "Typewords": ["nullable", "string"] }] },
		"Provider_porkbun": { "Name": "Provider_porkbun", "Docs": "", "Fields": [{ "Name": "api_key", "Docs": "", "Typewords": ["nullable", "string"] }, { "Name": "api_secret_key", "Docs": "", "Typewords": ["nullable", "string"] }] },
		"Provider_powerdns": { "Name": "Provider_powerdns", "Docs": "", "Fields": [{ "Name": "server_url", "Docs": "", "Typewords": ["string"] }, { "Name": "server_id", "Docs": "", "Typewords": ["nullable", "string"] }, { "Name": "api_token", "Docs": "", "Typewords": ["nullable", "string"] }, { "Name": "debug", "Docs": "", "Typewords": ["nullable", "string"] }] },
		"Provider_rfc2136": { "Name": "Provider_rfc2136", "Docs": "", "Fields": [{ "Name": "key_name", "Docs": "", "Typewords": ["nullable", "string"] }, { "Name": "key_alg", "Docs": "", "Typewords": ["nullable", "string"] }, { "Name": "key", "Docs": "", "Typewords": ["nullable", "string"] }, { "Name": "server", "Docs": "", "Typewords": ["nullable", "string"] }] },
		"Provider_route53": { "Name": "Provider_route53", "Docs": "", "Fields": [{ "Name": "region", "Docs": "", "Typewords": ["nullable", "string"] }, { "Name": "aws_profile", "Docs": "", "Typewords": ["nullable", "string"] }, { "Name": "profile", "Docs": "", "Typewords": ["nullable", "string"] }, { "Name": "access_key_id", "Docs": "", "Typewords": ["nullable", "string"] }, { "Name": "secret_access_key", "Docs": "", "Typewords": ["nullable", "string"] }, { "Name": "token", "Docs": "", "Typewords": ["nullable", "string"] }, { "Name": "session_token", "Docs": "", "Typewords": ["nullable", "string"] }, { "Name": "max_retries", "Docs": "", "Typewords": ["nullable", "int32"] }, { "Name": "max_wait_dur", "Docs": "", "Typewords": ["nullable", "int64"] }, { "Name": "wait_for_propagation", "Docs": "", "Typewords": ["nullable", "bool"] }, { "Name": "hosted_zone_id", "Docs": "", "Typewords": ["nullable", "string"] }] },
		"Provider_scaleway": { "Name": "Provider_scaleway", "Docs": "", "Fields": [{ "Name": "secret_key", "Docs": "", "Typewords": ["nullable", "string"] }, { "Name": "organization_id", "Docs": "", "Typewords": ["nullable", "string"] }] },
		"Provider_selectel": { "Name": "Provider_selectel", "Docs": "", "Fields": [{ "Name": "user", "Docs": "", "Typewords": ["nullable", "string"] }, { "Name": "password", "Docs": "", "Typewords": ["nullable", "string"] }, { "Name": "account_id", "Docs": "", "Typewords": ["nullable", "string"] }, { "Name": "project_name", "Docs": "", "Typewords": ["nullable", "string"] }, { "Name": "KeystoneToken", "Docs": "", "Typewords": ["string"] }] },
		"Provider_tencentcloud": { "Name": "Provider_tencentcloud", "Docs": "", "Fields": [{ "Name": "SecretId", "Docs": "", "Typewords": ["string"] }, { "Name": "SecretKey", "Docs": "", "Typewords": ["string"] }] },
		"Provider_timeweb": { "Name": "Provider_timeweb", "Docs": "", "Fields": [{ "Name": "ApiURL", "Docs": "", "Typewords": ["string"] }, { "Name": "ApiToken", "Docs": "", "Typewords": ["string"] }] },
		"Provider_totaluptime": { "Name": "Provider_totaluptime", "Docs": "", "Fields": [{ "Name": "username", "Docs": "", "Typewords": ["nullable", "string"] }, { "Name": "password", "Docs": "", "Typewords": ["nullable", "string"] }] },
		"Provider_vultr": { "Name": "Provider_vultr", "Docs": "", "Fields": [{ "Name": "api_token", "Docs": "", "Typewords": ["nullable", "string"] }] },
		"sherpadocSection": { "Name": "sherpadocSection", "Docs": "", "Fields": [{ "Name": "Name", "Docs": "", "Typewords": ["string"] }, { "Name": "Docs", "Docs": "", "Typewords": ["string"] }, { "Name": "Functions", "Docs": "", "Typewords": ["[]", "nullable", "sherpadocFunction"] }, { "Name": "Sections", "Docs": "", "Typewords": ["[]", "nullable", "sherpadocSection"] }, { "Name": "Structs", "Docs": "", "Typewords": ["[]", "sherpadocStruct"] }, { "Name": "Ints", "Docs": "", "Typewords": ["[]", "sherpadocInts"] }, { "Name": "Strings", "Docs": "", "Typewords": ["[]", "sherpadocStrings"] }, { "Name": "Version", "Docs": "", "Typewords": ["nullable", "string"] }, { "Name": "SherpaVersion", "Docs": "", "Typewords": ["int32"] }, { "Name": "SherpadocVersion", "Docs": "", "Typewords": ["nullable", "int32"] }] },
		"sherpadocFunction": { "Name": "sherpadocFunction", "Docs": "", "Fields": [{ "Name": "Name", "Docs": "", "Typewords": ["string"] }, { "Name": "Docs", "Docs": "", "Typewords": ["string"] }, { "Name": "Params", "Docs": "", "Typewords": ["[]", "sherpadocArg"] }, { "Name": "Returns", "Docs": "", "Typewords": ["[]", "sherpadocArg"] }] },
		"sherpadocArg": { "Name": "sherpadocArg", "Docs": "", "Fields": [{ "Name": "Name", "Docs": "", "Typewords": ["string"] }, { "Name": "Typewords", "Docs": "", "Typewords": ["[]", "string"] }] },
		"sherpadocStruct": { "Name": "sherpadocStruct", "Docs": "", "Fields": [{ "Name": "Name", "Docs": "", "Typewords": ["string"] }, { "Name": "Docs", "Docs": "", "Typewords": ["string"] }, { "Name": "Fields", "Docs": "", "Typewords": ["[]", "sherpadocField"] }] },
		"sherpadocField": { "Name": "sherpadocField", "Docs": "", "Fields": [{ "Name": "Name", "Docs": "", "Typewords": ["string"] }, { "Name": "Docs", "Docs": "", "Typewords": ["string"] }, { "Name": "Typewords", "Docs": "", "Typewords": ["[]", "string"] }] },
		"sherpadocInts": { "Name": "sherpadocInts", "Docs": "", "Fields": [{ "Name": "Name", "Docs": "", "Typewords": ["string"] }, { "Name": "Docs", "Docs": "", "Typewords": ["string"] }, { "Name": "Values", "Docs": "", "Typewords": ["[]", "IntValue"] }] },
		"IntValue": { "Name": "IntValue", "Docs": "", "Fields": [{ "Name": "Name", "Docs": "", "Typewords": ["string"] }, { "Name": "Value", "Docs": "", "Typewords": ["int64"] }, { "Name": "Docs", "Docs": "", "Typewords": ["string"] }] },
		"sherpadocStrings": { "Name": "sherpadocStrings", "Docs": "", "Fields": [{ "Name": "Name", "Docs": "", "Typewords": ["string"] }, { "Name": "Docs", "Docs": "", "Typewords": ["string"] }, { "Name": "Values", "Docs": "", "Typewords": ["[]", "StringValue"] }] },
		"StringValue": { "Name": "StringValue", "Docs": "", "Fields": [{ "Name": "Name", "Docs": "", "Typewords": ["string"] }, { "Name": "Value", "Docs": "", "Typewords": ["string"] }, { "Name": "Docs", "Docs": "", "Typewords": ["string"] }] },
		"BaseURL": { "Name": "BaseURL", "Docs": "", "Values": [{ "Name": "Sandbox", "Value": "https://api.sandbox.dnsmadeeasy.com/V2.0/", "Docs": "" }, { "Name": "Prod", "Value": "https://api.dnsmadeeasy.com/V2.0/", "Docs": "" }] },
	};
	api.parser = {
		Zone: (v) => api.parse("Zone", v),
		ProviderConfig: (v) => api.parse("ProviderConfig", v),
		ZoneNotify: (v) => api.parse("ZoneNotify", v),
		Credential: (v) => api.parse("Credential", v),
		RecordSet: (v) => api.parse("RecordSet", v),
		Record: (v) => api.parse("Record", v),
		PropagationState: (v) => api.parse("PropagationState", v),
		RecordSetChange: (v) => api.parse("RecordSetChange", v),
		KnownProviders: (v) => api.parse("KnownProviders", v),
		Provider_alidns: (v) => api.parse("Provider_alidns", v),
		Provider_autodns: (v) => api.parse("Provider_autodns", v),
		Provider_azure: (v) => api.parse("Provider_azure", v),
		Provider_bunny: (v) => api.parse("Provider_bunny", v),
		Provider_civo: (v) => api.parse("Provider_civo", v),
		Provider_cloudflare: (v) => api.parse("Provider_cloudflare", v),
		Provider_cloudns: (v) => api.parse("Provider_cloudns", v),
		Provider_ddnss: (v) => api.parse("Provider_ddnss", v),
		Provider_desec: (v) => api.parse("Provider_desec", v),
		Provider_digitalocean: (v) => api.parse("Provider_digitalocean", v),
		Provider_directadmin: (v) => api.parse("Provider_directadmin", v),
		Provider_dnsimple: (v) => api.parse("Provider_dnsimple", v),
		Provider_dnsmadeeasy: (v) => api.parse("Provider_dnsmadeeasy", v),
		Provider_dnspod: (v) => api.parse("Provider_dnspod", v),
		Provider_dnsupdate: (v) => api.parse("Provider_dnsupdate", v),
		Provider_domainnameshop: (v) => api.parse("Provider_domainnameshop", v),
		Provider_dreamhost: (v) => api.parse("Provider_dreamhost", v),
		Provider_duckdns: (v) => api.parse("Provider_duckdns", v),
		Provider_dynu: (v) => api.parse("Provider_dynu", v),
		Provider_dynv6: (v) => api.parse("Provider_dynv6", v),
		Provider_easydns: (v) => api.parse("Provider_easydns", v),
		Provider_exoscale: (v) => api.parse("Provider_exoscale", v),
		Provider_gandi: (v) => api.parse("Provider_gandi", v),
		Provider_gcore: (v) => api.parse("Provider_gcore", v),
		Provider_glesys: (v) => api.parse("Provider_glesys", v),
		Provider_godaddy: (v) => api.parse("Provider_godaddy", v),
		Provider_googleclouddns: (v) => api.parse("Provider_googleclouddns", v),
		Provider_he: (v) => api.parse("Provider_he", v),
		Provider_hetzner: (v) => api.parse("Provider_hetzner", v),
		Provider_hexonet: (v) => api.parse("Provider_hexonet", v),
		Provider_hosttech: (v) => api.parse("Provider_hosttech", v),
		Provider_huaweicloud: (v) => api.parse("Provider_huaweicloud", v),
		Provider_infomaniak: (v) => api.parse("Provider_infomaniak", v),
		Provider_inwx: (v) => api.parse("Provider_inwx", v),
		Provider_ionos: (v) => api.parse("Provider_ionos", v),
		Provider_katapult: (v) => api.parse("Provider_katapult", v),
		Provider_leaseweb: (v) => api.parse("Provider_leaseweb", v),
		Provider_linode: (v) => api.parse("Provider_linode", v),
		Provider_loopia: (v) => api.parse("Provider_loopia", v),
		Provider_luadns: (v) => api.parse("Provider_luadns", v),
		Provider_mailinabox: (v) => api.parse("Provider_mailinabox", v),
		Provider_metaname: (v) => api.parse("Provider_metaname", v),
		Provider_mythicbeasts: (v) => api.parse("Provider_mythicbeasts", v),
		Provider_namecheap: (v) => api.parse("Provider_namecheap", v),
		Provider_namedotcom: (v) => api.parse("Provider_namedotcom", v),
		Provider_namesilo: (v) => api.parse("Provider_namesilo", v),
		Provider_nanelo: (v) => api.parse("Provider_nanelo", v),
		Provider_netcup: (v) => api.parse("Provider_netcup", v),
		Provider_netlify: (v) => api.parse("Provider_netlify", v),
		Provider_nfsn: (v) => api.parse("Provider_nfsn", v),
		Provider_njalla: (v) => api.parse("Provider_njalla", v),
		Provider: (v) => api.parse("Provider", v),
		AuthOpenStack: (v) => api.parse("AuthOpenStack", v),
		Provider_ovh: (v) => api.parse("Provider_ovh", v),
		Provider_porkbun: (v) => api.parse("Provider_porkbun", v),
		Provider_powerdns: (v) => api.parse("Provider_powerdns", v),
		Provider_rfc2136: (v) => api.parse("Provider_rfc2136", v),
		Provider_route53: (v) => api.parse("Provider_route53", v),
		Provider_scaleway: (v) => api.parse("Provider_scaleway", v),
		Provider_selectel: (v) => api.parse("Provider_selectel", v),
		Provider_tencentcloud: (v) => api.parse("Provider_tencentcloud", v),
		Provider_timeweb: (v) => api.parse("Provider_timeweb", v),
		Provider_totaluptime: (v) => api.parse("Provider_totaluptime", v),
		Provider_vultr: (v) => api.parse("Provider_vultr", v),
		sherpadocSection: (v) => api.parse("sherpadocSection", v),
		sherpadocFunction: (v) => api.parse("sherpadocFunction", v),
		sherpadocArg: (v) => api.parse("sherpadocArg", v),
		sherpadocStruct: (v) => api.parse("sherpadocStruct", v),
		sherpadocField: (v) => api.parse("sherpadocField", v),
		sherpadocInts: (v) => api.parse("sherpadocInts", v),
		IntValue: (v) => api.parse("IntValue", v),
		sherpadocStrings: (v) => api.parse("sherpadocStrings", v),
		StringValue: (v) => api.parse("StringValue", v),
		BaseURL: (v) => api.parse("BaseURL", v),
	};
	// API is the webapi used by the admin frontend.
	let defaultOptions = { slicesNullable: true, mapsNullable: true, nullableOptional: true };
	class Client {
		baseURL;
		authState;
		options;
		constructor() {
			this.authState = {};
			this.options = { ...defaultOptions };
			this.baseURL = this.options.baseURL || api.defaultBaseURL;
		}
		withAuthToken(token) {
			const c = new Client();
			c.authState.token = token;
			c.options = this.options;
			return c;
		}
		withOptions(options) {
			const c = new Client();
			c.authState = this.authState;
			c.options = { ...this.options, ...options };
			return c;
		}
		// Zones returns all zones.
		async Zones() {
			const fn = "Zones";
			const paramTypes = [];
			const returnTypes = [["[]", "Zone"]];
			const params = [];
			return await _sherpaCall(this.baseURL, this.authState, { ...this.options }, paramTypes, returnTypes, fn, params);
		}
		// Zone returns details about a single zone, the provider config, dns notify
		// destinations, credentials with access to the zone, and record sets. The returned
		// record sets includes those no long active (i.e. deleted). The
		// history/propagation state fo the record sets only includes those that may still
		// be in caches. Use ZoneRecordSetHistory for the full history for a single record
		// set.
		async Zone(zone) {
			const fn = "Zone";
			const paramTypes = [["string"]];
			const returnTypes = [["Zone"], ["ProviderConfig"], ["[]", "ZoneNotify"], ["[]", "Credential"], ["[]", "RecordSet"]];
			const params = [zone];
			return await _sherpaCall(this.baseURL, this.authState, { ...this.options }, paramTypes, returnTypes, fn, params);
		}
		// ZoneRecords returns all records for a zone, including historic records, without
		// grouping them into record sets.
		async ZoneRecords(zone) {
			const fn = "ZoneRecords";
			const paramTypes = [["string"]];
			const returnTypes = [["[]", "Record"]];
			const params = [zone];
			return await _sherpaCall(this.baseURL, this.authState, { ...this.options }, paramTypes, returnTypes, fn, params);
		}
		// ZoneRefresh starts a sync of the records from the provider into the local
		// database, sending dns notify if needed. ZoneRefresh returns all records
		// (included deleted) from after the synchronization.
		async ZoneRefresh(zone) {
			const fn = "ZoneRefresh";
			const paramTypes = [["string"]];
			const returnTypes = [["Zone"], ["[]", "RecordSet"]];
			const params = [zone];
			return await _sherpaCall(this.baseURL, this.authState, { ...this.options }, paramTypes, returnTypes, fn, params);
		}
		// ZonePurgeHistory removes historic records from the database, those marked "deleted".
		async ZonePurgeHistory(zone) {
			const fn = "ZonePurgeHistory";
			const paramTypes = [["string"]];
			const returnTypes = [["Zone"], ["[]", "RecordSet"]];
			const params = [zone];
			return await _sherpaCall(this.baseURL, this.authState, { ...this.options }, paramTypes, returnTypes, fn, params);
		}
		// ZoneAdd adds a new zone to the database. A TSIG credential is created
		// automatically. Records are fetched returning the new zone, in the background.
		// 
		// If pc.ProviderName is non-empty, a new ProviderConfig is added.
		async ZoneAdd(z, notifies) {
			const fn = "ZoneAdd";
			const paramTypes = [["Zone"], ["[]", "ZoneNotify"]];
			const returnTypes = [["Zone"]];
			const params = [z, notifies];
			return await _sherpaCall(this.baseURL, this.authState, { ...this.options }, paramTypes, returnTypes, fn, params);
		}
		// ZoneDelete removes a zone and all its records, credentials and dns notify addresses, from the database.
		async ZoneDelete(zone) {
			const fn = "ZoneDelete";
			const paramTypes = [["string"]];
			const returnTypes = [];
			const params = [zone];
			return await _sherpaCall(this.baseURL, this.authState, { ...this.options }, paramTypes, returnTypes, fn, params);
		}
		// ZoneUpdate updates the provider config and refresh & sync interval for a zone.
		async ZoneUpdate(z) {
			const fn = "ZoneUpdate";
			const paramTypes = [["Zone"]];
			const returnTypes = [["Zone"]];
			const params = [z];
			return await _sherpaCall(this.baseURL, this.authState, { ...this.options }, paramTypes, returnTypes, fn, params);
		}
		// ZoneNotify send a DNS notify message to an address.
		async ZoneNotify(zoneNotifyID) {
			const fn = "ZoneNotify";
			const paramTypes = [["int64"]];
			const returnTypes = [];
			const params = [zoneNotifyID];
			return await _sherpaCall(this.baseURL, this.authState, { ...this.options }, paramTypes, returnTypes, fn, params);
		}
		// ZoneNotifyAdd adds a new DNS NOTIFY destination to a zone.
		async ZoneNotifyAdd(zn) {
			const fn = "ZoneNotifyAdd";
			const paramTypes = [["ZoneNotify"]];
			const returnTypes = [["ZoneNotify"]];
			const params = [zn];
			return await _sherpaCall(this.baseURL, this.authState, { ...this.options }, paramTypes, returnTypes, fn, params);
		}
		// ZoneNotifyDelete removes a DNS NOTIFY destination from a zone.
		async ZoneNotifyDelete(zoneNotifyID) {
			const fn = "ZoneNotifyDelete";
			const paramTypes = [["int64"]];
			const returnTypes = [];
			const params = [zoneNotifyID];
			return await _sherpaCall(this.baseURL, this.authState, { ...this.options }, paramTypes, returnTypes, fn, params);
		}
		// ZoneCredentialAdd adds a new TSIG or TLS public key credential to a zone.
		async ZoneCredentialAdd(zone, c) {
			const fn = "ZoneCredentialAdd";
			const paramTypes = [["string"], ["Credential"]];
			const returnTypes = [["Credential"]];
			const params = [zone, c];
			return await _sherpaCall(this.baseURL, this.authState, { ...this.options }, paramTypes, returnTypes, fn, params);
		}
		// ZoneCredentialDelete removes a TSIG/TLS public key credential from a zone.
		async ZoneCredentialDelete(credentialID) {
			const fn = "ZoneCredentialDelete";
			const paramTypes = [["int64"]];
			const returnTypes = [];
			const params = [credentialID];
			return await _sherpaCall(this.baseURL, this.authState, { ...this.options }, paramTypes, returnTypes, fn, params);
		}
		// ZoneImportRecords parses records in zonefile, assuming standard zone file syntax,
		// and adds the records via the provider and syncs the newly added records to the
		// local database. The latest records, included historic/deleted records after the
		// sync are returned.
		async ZoneImportRecords(zone, zonefile) {
			const fn = "ZoneImportRecords";
			const paramTypes = [["string"], ["string"]];
			const returnTypes = [["[]", "Record"]];
			const params = [zone, zonefile];
			return await _sherpaCall(this.baseURL, this.authState, { ...this.options }, paramTypes, returnTypes, fn, params);
		}
		// RecordSetAdd adds a record set through the provider, then waits for it to
		// synchronize back to the local database.
		// 
		// The name and type must not already exist. Use RecordSetUpdate to add values to
		// an existing record set.
		// 
		// The inserted records are returned.
		async RecordSetAdd(zone, rsc) {
			const fn = "RecordSetAdd";
			const paramTypes = [["string"], ["RecordSetChange"]];
			const returnTypes = [["[]", "Record"]];
			const params = [zone, rsc];
			return await _sherpaCall(this.baseURL, this.authState, { ...this.options }, paramTypes, returnTypes, fn, params);
		}
		// RecordSetUpdate updates an existing record set, replacing its values with the
		// new values. If the name has changed, the old records are deleted and new records
		// with new name inserted.
		// 
		// Before changing, prevRecordIDs are compared with the current records for the
		// name and type, and must be the same.
		// 
		// valueRecordIDs match Values from RecordNewSet (must have the same number of
		// items). New values must have 0 as record ID.
		// 
		// The records of the updated record set are returned.
		async RecordSetUpdate(zone, oldRelName, rsc, prevRecordIDs, valueRecordIDs) {
			const fn = "RecordSetUpdate";
			const paramTypes = [["string"], ["string"], ["RecordSetChange"], ["[]", "int64"], ["[]", "int64"]];
			const returnTypes = [["[]", "Record"]];
			const params = [zone, oldRelName, rsc, prevRecordIDs, valueRecordIDs];
			return await _sherpaCall(this.baseURL, this.authState, { ...this.options }, paramTypes, returnTypes, fn, params);
		}
		// RecordSetDelete removes a record set through the provider and waits for the
		// change to be synced to the local database. The historic/deleted record is
		// returned.
		// 
		// recordIDs must be the current record ids the caller expects to invalidate.
		// 
		// The updated records, now marked as deleted, are returned.
		async RecordSetDelete(zone, relName, typ, recordIDs) {
			const fn = "RecordSetDelete";
			const paramTypes = [["string"], ["string"], ["uint16"], ["[]", "int64"]];
			const returnTypes = [["[]", "Record"]];
			const params = [zone, relName, typ, recordIDs];
			return await _sherpaCall(this.baseURL, this.authState, { ...this.options }, paramTypes, returnTypes, fn, params);
		}
		// Version returns the version of this build of the application.
		async Version() {
			const fn = "Version";
			const paramTypes = [];
			const returnTypes = [["string"]];
			const params = [];
			return await _sherpaCall(this.baseURL, this.authState, { ...this.options }, paramTypes, returnTypes, fn, params);
		}
		// DNSTypeNames returns a mapping of DNS type numbers to strings.
		async DNSTypeNames() {
			const fn = "DNSTypeNames";
			const paramTypes = [];
			const returnTypes = [["{}", "string"]];
			const params = [];
			return await _sherpaCall(this.baseURL, this.authState, { ...this.options }, paramTypes, returnTypes, fn, params);
		}
		// KnownProviders is a dummy method whose sole purpose is to get an API description
		// of all known providers in the API documentation, for use in TypeScript.
		async KnownProviders() {
			const fn = "KnownProviders";
			const paramTypes = [];
			const returnTypes = [["KnownProviders"], ["sherpadocSection"]];
			const params = [];
			return await _sherpaCall(this.baseURL, this.authState, { ...this.options }, paramTypes, returnTypes, fn, params);
		}
		// Docs returns the API docs. The TypeScript code uses this documentation to build
		// a UI for the fields in configurations for providers (as included through
		// KnownProviders).
		async Docs() {
			const fn = "Docs";
			const paramTypes = [];
			const returnTypes = [["sherpadocSection"]];
			const params = [];
			return await _sherpaCall(this.baseURL, this.authState, { ...this.options }, paramTypes, returnTypes, fn, params);
		}
		// ProviderConfigTest tests the provider configuration for zone. Used before
		// creating a zone with a new config or updating the config for an existing zone.
		async ProviderConfigTest(zone, provider, providerConfigJSON) {
			const fn = "ProviderConfigTest";
			const paramTypes = [["string"], ["string"], ["string"]];
			const returnTypes = [["int32"]];
			const params = [zone, provider, providerConfigJSON];
			return await _sherpaCall(this.baseURL, this.authState, { ...this.options }, paramTypes, returnTypes, fn, params);
		}
		// ProviderConfigs returns all provider configs.
		async ProviderConfigs() {
			const fn = "ProviderConfigs";
			const paramTypes = [];
			const returnTypes = [["[]", "ProviderConfig"]];
			const params = [];
			return await _sherpaCall(this.baseURL, this.authState, { ...this.options }, paramTypes, returnTypes, fn, params);
		}
		// ProviderURLs returns a mapping of provider names to URLs of their
		// repositories, for further help/instructions.
		async ProviderURLs() {
			const fn = "ProviderURLs";
			const paramTypes = [];
			const returnTypes = [["{}", "string"]];
			const params = [];
			return await _sherpaCall(this.baseURL, this.authState, { ...this.options }, paramTypes, returnTypes, fn, params);
		}
		// ProviderConfigAdd adds a new provider config.
		async ProviderConfigAdd(pc) {
			const fn = "ProviderConfigAdd";
			const paramTypes = [["ProviderConfig"]];
			const returnTypes = [["ProviderConfig"]];
			const params = [pc];
			return await _sherpaCall(this.baseURL, this.authState, { ...this.options }, paramTypes, returnTypes, fn, params);
		}
		// ProviderConfigUpdate updates a provider config.
		async ProviderConfigUpdate(pc) {
			const fn = "ProviderConfigUpdate";
			const paramTypes = [["ProviderConfig"]];
			const returnTypes = [["ProviderConfig"]];
			const params = [pc];
			return await _sherpaCall(this.baseURL, this.authState, { ...this.options }, paramTypes, returnTypes, fn, params);
		}
		// ZoneRecordSets returns the current record sets including propagation states that
		// are not the latest version but that may still be in caches. For the full history
		// of a record set, see ZoneRecordSetHistory.
		async ZoneRecordSets(zone) {
			const fn = "ZoneRecordSets";
			const paramTypes = [["string"]];
			const returnTypes = [["[]", "RecordSet"]];
			const params = [zone];
			return await _sherpaCall(this.baseURL, this.authState, { ...this.options }, paramTypes, returnTypes, fn, params);
		}
		// ZoneRecordSetHistory returns the propagation state history for a record set,
		// including the current value.
		async ZoneRecordSetHistory(zone, relName, typ) {
			const fn = "ZoneRecordSetHistory";
			const paramTypes = [["string"], ["string"], ["uint16"]];
			const returnTypes = [["[]", "PropagationState"]];
			const params = [zone, relName, typ];
			return await _sherpaCall(this.baseURL, this.authState, { ...this.options }, paramTypes, returnTypes, fn, params);
		}
	}
	api.Client = Client;
	api.defaultBaseURL = (function () {
		let p = location.pathname;
		if (p && p[p.length - 1] !== '/') {
			let l = location.pathname.split('/');
			l = l.slice(0, l.length - 1);
			p = '/' + l.join('/') + '/';
		}
		return location.protocol + '//' + location.host + p + 'api/';
	})();
	// NOTE: code below is shared between github.com/mjl-/sherpaweb and github.com/mjl-/sherpats.
	// KEEP IN SYNC.
	api.supportedSherpaVersion = 1;
	// verifyArg typechecks "v" against "typewords", returning a new (possibly modified) value for JSON-encoding.
	// toJS indicate if the data is coming into JS. If so, timestamps are turned into JS Dates. Otherwise, JS Dates are turned into strings.
	// allowUnknownKeys configures whether unknown keys in structs are allowed.
	// types are the named types of the API.
	api.verifyArg = (path, v, typewords, toJS, allowUnknownKeys, types, opts) => {
		return new verifier(types, toJS, allowUnknownKeys, opts).verify(path, v, typewords);
	};
	api.parse = (name, v) => api.verifyArg(name, v, [name], true, false, api.types, defaultOptions);
	class verifier {
		types;
		toJS;
		allowUnknownKeys;
		opts;
		constructor(types, toJS, allowUnknownKeys, opts) {
			this.types = types;
			this.toJS = toJS;
			this.allowUnknownKeys = allowUnknownKeys;
			this.opts = opts;
		}
		verify(path, v, typewords) {
			typewords = typewords.slice(0);
			const ww = typewords.shift();
			const error = (msg) => {
				if (path != '') {
					msg = path + ': ' + msg;
				}
				throw new Error(msg);
			};
			if (typeof ww !== 'string') {
				error('bad typewords');
				return; // should not be necessary, typescript doesn't see error always throws an exception?
			}
			const w = ww;
			const ensure = (ok, expect) => {
				if (!ok) {
					error('got ' + JSON.stringify(v) + ', expected ' + expect);
				}
				return v;
			};
			switch (w) {
				case 'nullable':
					if (v === null || v === undefined && this.opts.nullableOptional) {
						return v;
					}
					return this.verify(path, v, typewords);
				case '[]':
					if (v === null && this.opts.slicesNullable || v === undefined && this.opts.slicesNullable && this.opts.nullableOptional) {
						return v;
					}
					ensure(Array.isArray(v), "array");
					return v.map((e, i) => this.verify(path + '[' + i + ']', e, typewords));
				case '{}':
					if (v === null && this.opts.mapsNullable || v === undefined && this.opts.mapsNullable && this.opts.nullableOptional) {
						return v;
					}
					ensure(v !== null || typeof v === 'object', "object");
					const r = {};
					for (const k in v) {
						r[k] = this.verify(path + '.' + k, v[k], typewords);
					}
					return r;
			}
			ensure(typewords.length == 0, "empty typewords");
			const t = typeof v;
			switch (w) {
				case 'any':
					return v;
				case 'bool':
					ensure(t === 'boolean', 'bool');
					return v;
				case 'int8':
				case 'uint8':
				case 'int16':
				case 'uint16':
				case 'int32':
				case 'uint32':
				case 'int64':
				case 'uint64':
					ensure(t === 'number' && Number.isInteger(v), 'integer');
					return v;
				case 'float32':
				case 'float64':
					ensure(t === 'number', 'float');
					return v;
				case 'int64s':
				case 'uint64s':
					ensure(t === 'number' && Number.isInteger(v) || t === 'string', 'integer fitting in float without precision loss, or string');
					return '' + v;
				case 'string':
					ensure(t === 'string', 'string');
					return v;
				case 'timestamp':
					if (this.toJS) {
						ensure(t === 'string', 'string, with timestamp');
						const d = new Date(v);
						if (d instanceof Date && !isNaN(d.getTime())) {
							return d;
						}
						error('invalid date ' + v);
					}
					else {
						ensure(t === 'object' && v !== null, 'non-null object');
						ensure(v.__proto__ === Date.prototype, 'Date');
						return v.toISOString();
					}
			}
			// We're left with named types.
			const nt = this.types[w];
			if (!nt) {
				error('unknown type ' + w);
			}
			if (v === null) {
				error('bad value ' + v + ' for named type ' + w);
			}
			if (api.structTypes[nt.Name]) {
				const t = nt;
				if (typeof v !== 'object') {
					error('bad value ' + v + ' for struct ' + w);
				}
				const r = {};
				for (const f of t.Fields) {
					r[f.Name] = this.verify(path + '.' + f.Name, v[f.Name], f.Typewords);
				}
				// If going to JSON also verify no unknown fields are present.
				if (!this.allowUnknownKeys) {
					const known = {};
					for (const f of t.Fields) {
						known[f.Name] = true;
					}
					Object.keys(v).forEach((k) => {
						if (!known[k]) {
							error('unknown key ' + k + ' for struct ' + w);
						}
					});
				}
				return r;
			}
			else if (api.stringsTypes[nt.Name]) {
				const t = nt;
				if (typeof v !== 'string') {
					error('mistyped value ' + v + ' for named strings ' + t.Name);
				}
				if (!t.Values || t.Values.length === 0) {
					return v;
				}
				for (const sv of t.Values) {
					if (sv.Value === v) {
						return v;
					}
				}
				error('unknown value ' + v + ' for named strings ' + t.Name);
			}
			else if (api.intsTypes[nt.Name]) {
				const t = nt;
				if (typeof v !== 'number' || !Number.isInteger(v)) {
					error('mistyped value ' + v + ' for named ints ' + t.Name);
				}
				if (!t.Values || t.Values.length === 0) {
					return v;
				}
				for (const sv of t.Values) {
					if (sv.Value === v) {
						return v;
					}
				}
				error('unknown value ' + v + ' for named ints ' + t.Name);
			}
			else {
				throw new Error('unexpected named type ' + nt);
			}
		}
	}
	const _sherpaCall = async (baseURL, authState, options, paramTypes, returnTypes, name, params) => {
		if (!options.skipParamCheck) {
			if (params.length !== paramTypes.length) {
				return Promise.reject({ message: 'wrong number of parameters in sherpa call, saw ' + params.length + ' != expected ' + paramTypes.length });
			}
			params = params.map((v, index) => api.verifyArg('params[' + index + ']', v, paramTypes[index], false, false, api.types, options));
		}
		const simulate = async (json) => {
			const config = JSON.parse(json || 'null') || {};
			const waitMinMsec = config.waitMinMsec || 0;
			const waitMaxMsec = config.waitMaxMsec || 0;
			const wait = Math.random() * (waitMaxMsec - waitMinMsec);
			const failRate = config.failRate || 0;
			return new Promise((resolve, reject) => {
				if (options.aborter) {
					options.aborter.abort = () => {
						reject({ message: 'call to ' + name + ' aborted by user', code: 'sherpa:aborted' });
						reject = resolve = () => { };
					};
				}
				setTimeout(() => {
					const r = Math.random();
					if (r < failRate) {
						reject({ message: 'injected failure on ' + name, code: 'server:injected' });
					}
					else {
						resolve();
					}
					reject = resolve = () => { };
				}, waitMinMsec + wait);
			});
		};
		// Only simulate when there is a debug string. Otherwise it would always interfere
		// with setting options.aborter.
		let json = '';
		try {
			json = window.localStorage.getItem('sherpats-debug') || '';
		}
		catch (err) { }
		if (json) {
			await simulate(json);
		}
		const fn = (resolve, reject) => {
			let resolve1 = (v) => {
				resolve(v);
				resolve1 = () => { };
				reject1 = () => { };
			};
			let reject1 = (v) => {
				if ((v.code === 'user:noAuth' || v.code === 'user:badAuth') && options.login) {
					const login = options.login;
					if (!authState.loginPromise) {
						authState.loginPromise = new Promise((aresolve, areject) => {
							login(v.code === 'user:badAuth' ? (v.message || '') : '')
								.then((token) => {
								authState.token = token;
								authState.loginPromise = undefined;
								aresolve();
							}, (err) => {
								authState.loginPromise = undefined;
								areject(err);
							});
						});
					}
					authState.loginPromise
						.then(() => {
						fn(resolve, reject);
					}, (err) => {
						reject(err);
					});
					return;
				}
				reject(v);
				resolve1 = () => { };
				reject1 = () => { };
			};
			const url = baseURL + name;
			const req = new window.XMLHttpRequest();
			if (options.aborter) {
				options.aborter.abort = () => {
					req.abort();
					reject1({ code: 'sherpa:aborted', message: 'request aborted' });
				};
			}
			req.open('POST', url, true);
			if (options.csrfHeader && authState.token) {
				req.setRequestHeader(options.csrfHeader, authState.token);
			}
			if (options.timeoutMsec) {
				req.timeout = options.timeoutMsec;
			}
			req.onload = () => {
				if (req.status !== 200) {
					if (req.status === 404) {
						reject1({ code: 'sherpa:badFunction', message: 'function does not exist' });
					}
					else {
						reject1({ code: 'sherpa:http', message: 'error calling function, HTTP status: ' + req.status });
					}
					return;
				}
				let resp;
				try {
					resp = JSON.parse(req.responseText);
				}
				catch (err) {
					reject1({ code: 'sherpa:badResponse', message: 'bad JSON from server' });
					return;
				}
				if (resp && resp.error) {
					const err = resp.error;
					reject1({ code: err.code, message: err.message });
					return;
				}
				else if (!resp || !resp.hasOwnProperty('result')) {
					reject1({ code: 'sherpa:badResponse', message: "invalid sherpa response object, missing 'result'" });
					return;
				}
				if (options.skipReturnCheck) {
					resolve1(resp.result);
					return;
				}
				let result = resp.result;
				try {
					if (returnTypes.length === 0) {
						if (result) {
							throw new Error('function ' + name + ' returned a value while prototype says it returns "void"');
						}
					}
					else if (returnTypes.length === 1) {
						result = api.verifyArg('result', result, returnTypes[0], true, true, api.types, options);
					}
					else {
						if (result.length != returnTypes.length) {
							throw new Error('wrong number of values returned by ' + name + ', saw ' + result.length + ' != expected ' + returnTypes.length);
						}
						result = result.map((v, index) => api.verifyArg('result[' + index + ']', v, returnTypes[index], true, true, api.types, options));
					}
				}
				catch (err) {
					let errmsg = 'bad types';
					if (err instanceof Error) {
						errmsg = err.message;
					}
					reject1({ code: 'sherpa:badTypes', message: errmsg });
				}
				resolve1(result);
			};
			req.onerror = () => {
				reject1({ code: 'sherpa:connection', message: 'connection failed' });
			};
			req.ontimeout = () => {
				reject1({ code: 'sherpa:timeout', message: 'request timeout' });
			};
			req.setRequestHeader('Content-Type', 'application/json');
			try {
				req.send(JSON.stringify({ params: params }));
			}
			catch (err) {
				reject1({ code: 'sherpa:badData', message: 'cannot marshal to JSON' });
			}
		};
		return await new Promise(fn);
	};
})(api || (api = {}));
let rootElem;
let crumbElem = dom.span();
let pageElem = dom.div(style({ padding: '1em' }), dom.div(style({ textAlign: 'center' }), 'Loading...'));
let version;
let dnsTypeNames = {};
const client = new api.Client();
const link = (href, anchor) => dom.a(attr.href(href), anchor);
const trimDot = (s) => s.replace(/\.$/, '');
const check = async (elem, fn) => {
	const overlay = dom.div(style({ position: 'fixed', top: 0, left: 0, right: 0, bottom: 0, zIndex: 2, backgroundColor: '#ffffff00' }));
	document.body.append(overlay);
	pageElem.classList.toggle('loading', true);
	if (elem) {
		elem.disabled = true;
	}
	try {
		const r = await fn();
		return r;
	}
	catch (err) {
		alert('Error: ' + err.message);
		throw err;
	}
	finally {
		overlay.remove();
		pageElem.classList.toggle('loading', false);
		if (elem) {
			elem.disabled = false;
		}
	}
};
const popupOpts = (opaque, ...kids) => {
	let close = () => { };
	const closed = new Promise((resolve, reject) => {
		const origFocus = document.activeElement;
		close = (canceled) => {
			if (!root.parentNode) {
				return;
			}
			root.remove();
			if (origFocus && origFocus instanceof HTMLElement && origFocus.parentNode) {
				origFocus.focus();
			}
			if (canceled) {
				reject();
			}
			else {
				resolve();
			}
		};
		let content;
		const root = dom.div(style({ position: 'fixed', top: 0, right: 0, bottom: 0, left: 0, paddingTop: '5vh', backgroundColor: opaque ? '#ffffff' : 'rgba(0, 0, 0, 0.1)', display: 'flex', alignItems: 'flex-start', justifyContent: 'center', zIndex: opaque ? 3 : 1 }), opaque ? [] : [
			function keydown(e) {
				if (e.key === 'Escape') {
					e.stopPropagation();
					close(true);
				}
			},
			function click(e) {
				e.stopPropagation();
				close(true);
			},
		], content = dom.div(attr.tabindex('0'), style({ backgroundColor: 'white', borderRadius: '.25em', padding: '1em', boxShadow: '0 0 20px rgba(0, 0, 0, 0.1)', border: '1px solid #ddd', maxWidth: '95vw', overflowX: 'auto', maxHeight: '90vh', overflowY: 'auto' }), function click(e) {
			e.stopPropagation();
		}, kids));
		document.body.appendChild(root);
		content.focus();
		return close;
	});
	return [close, closed];
};
const trimPrefix = (s, prefix) => s.startsWith(prefix) ? s.substring(prefix.length) : s;
const trimSuffix = (s, suffix) => s.endsWith(suffix) ? s.substring(0, s.length - suffix.length) : s;
const chunked = (l, len) => {
	const r = [];
	while (l.length > 0) {
		r.push(l.slice(0, len));
		l = l.slice(len);
	}
	return r;
};
const popup = (...kids) => popupOpts(false, ...kids);
const availableProviders = async () => {
	const docs = await client.Docs();
	const stringEnums = new Map();
	for (const e of (docs.Strings || [])) {
		stringEnums.set(e.Name, e);
	}
	const providers = (docs.Structs || []).filter(struct => struct.Name.startsWith('Provider_'));
	return [stringEnums, providers];
};
const providerConfigJSON = (fields) => {
	const config = {};
	for (const [k, f] of fields.fieldMap) {
		let v = null;
		if (f.nullable && !f.elem.value) {
		}
		else if (f.type === 'bool') {
			v = f.elem.checked;
		}
		else if (f.type === 'number') {
			v = parseInt(f.elem.value);
		}
		else {
			v = f.elem.value;
		}
		config[k] = v;
	}
	return JSON.stringify(config);
};
const providerFields = (p, stringEnums, configJSON) => {
	const fieldMap = new Map();
	let config = {};
	if (configJSON) {
		config = JSON.parse(configJSON);
	}
	const root = dom.div((p.Fields || []).map(f => {
		let tw = f.Typewords || [];
		let nullable = false;
		if (!tw[0]) {
			alert('missing type word');
			throw new Error('missing type word');
		}
		if (tw[0] === 'nullable') {
			tw = tw.slice(1);
		}
		if (!tw[0]) {
			alert('missing type word');
			throw new Error('missing type word');
		}
		if (tw[0] === 'bool') {
			const input = dom.input(attr.type('checkbox'), config[f.Name] === true ? attr.checked('') : []);
			fieldMap.set(f.Name, { elem: input, nullable: nullable, type: 'bool' });
			return dom.div(dom.div(style({ margin: '.5ex 0' }), dom.label(input, ' ', '"' + f.Name + '"')), f.Docs ? dom.div(style({ fontStyle: 'italic', maxWidth: '40em', marginBottom: '2ex' }), '"' + f.Docs + '"') : []);
		}
		let input = dom.input();
		let typ = 'string';
		if (tw[0] === 'int32' || tw[0] === 'int64' || tw[0] === 'uint32' || tw[1] === 'uint64') {
			input = dom.input(attr.type('number'), config[f.Name] ? attr.value('' + config[f.Name]) : []);
			typ = 'number';
		}
		else if (tw[0] === 'string') {
			input = dom.input(config[f.Name] ? attr.value('' + config[f.Name]) : []);
		}
		else {
			const values = stringEnums.get(tw[0]);
			if (values) {
				input = dom.select((values.Values || []).map(v => dom.option(`${v.Name} - ${v.Value}`, attr.value(v.Value), config[f.Name] === v.Value ? attr.selected('') : [])));
			}
			else {
				alert(`unknown type "${tw.join(' ')}" for field ${f.Name}`);
				input = dom.input();
			}
		}
		fieldMap.set(f.Name, { elem: input, nullable: nullable, type: typ });
		return dom.div(dom.div(dom.label('"' + f.Name + '"')), dom.div(style({ margin: '.5ex 0' }), input), f.Docs ? dom.div(style({ fontStyle: 'italic', maxWidth: '40em', marginBottom: '2ex' }), '"' + f.Docs + '"') : []);
	}));
	return { root: root, fieldMap: fieldMap };
};
const pageHome = async () => {
	let [zones0] = await Promise.all([
		client.Zones(),
	]);
	let zones = zones0 || [];
	dom._kids(crumbElem, dom.a(attr.href('#'), 'Home'));
	document.title = 'Dnsclay';
	let zonesTbody;
	const root = dom.div(dom.div(dom.clickbutton('Add zone', async function click() {
		let zone;
		let fieldset;
		let testResult;
		let newProviderConfigName;
		let existingProviderConfigName;
		const [[stringEnums, providers], providerURLs] = await Promise.all([
			availableProviders(),
			client.ProviderURLs(),
		]);
		const providerConfigs = await client.ProviderConfigs() || [];
		let fields;
		let providerName = '';
		const updateProviderConfig = () => {
			providerName = fieldset.querySelector('input[name=provider]:checked').value;
			const p = providers.find(p => p.Name === 'Provider_' + providerName);
			if (!p) {
				alert('cannot find provider ' + providerName);
				return;
			}
			const url = providerURLs[providerName];
			dom._kids(providerConfigBox, style({ display: 'flex', flexDirection: 'column', gap: '2ex' }), dom.label(dom.div('Name'), dom.div(newProviderConfigName = dom.input(attr.required(''), attr.value(newProviderConfigName?.value || zone.value)))), dom.div(style({ padding: '1em', border: '1px solid #ddd' }), dom.h2('"' + providerName + '" fields'), dom.p('Implemented through ', dom.a(attr.href('https://' + url), url, attr.rel('noreferrer noopener')), ', see ', dom.a(attr.href('https://pkg.go.dev/' + url), 'Go documentation', attr.rel('noreferrer noopener'))), dom.div(style({ display: 'flex', flexDirection: 'column', gap: '2ex' }), fields = providerFields(p, stringEnums, null))));
		};
		let providerConfigBox;
		const [close] = popup(dom.div(dom.h1('New zone'), dom.form(async function submit(e) {
			e.preventDefault();
			e.stopPropagation();
			// Test config.
			if (testResult) {
				testResult.innerText = '';
			}
			if (!zone.value) {
				alert('Zone required.');
				return;
			}
			let pName = '';
			let pcJSON = '';
			if (existingProviderConfigName.value) {
				const pc = providerConfigs.find(pc => pc.Name === existingProviderConfigName.value);
				if (!pc) {
					alert('Provider config not found.');
					return;
				}
				pName = pc.ProviderName;
				pcJSON = pc.ProviderConfigJSON;
			}
			else {
				if (!fields) {
					alert('No provider selected.');
					return;
				}
				pName = providerName;
				pcJSON = providerConfigJSON(fields);
			}
			const nrecords = await check(fieldset, () => client.ProviderConfigTest(trimSuffix(zone.value, '.') + '.', pName, pcJSON));
			testResult.innerText = 'Success, found ' + nrecords + ' DNS records';
		}, fieldset = dom.fieldset(style({ display: 'flex', flexDirection: 'column', gap: '2ex' }), dom.div(dom.div(dom.label('Zone')), zone = dom.input(attr.required(''))), dom.div(dom.div(dom.label('Create new provider config')), dom.div(style({ display: 'flex', gap: '1em' }), chunked(providers, 10).map(plist => dom.div(plist.map(p => {
			return dom.div(dom.label(dom.input(attr.type('radio'), attr.name('provider'), attr.value(trimPrefix(p.Name, 'Provider_')), function change() { updateProviderConfig(); }), ' ', trimPrefix(p.Name, 'Provider_')));
		}))))), providerConfigBox = dom.div(), dom.label(dom.div('Use existing provider config'), dom.div(existingProviderConfigName = dom.select(dom.option('', attr.value('')), providerConfigs.map(pc => dom.option(pc.Name))))), dom.div(dom.submitbutton('Test config'), ' ', testResult = dom.span()), dom.div(dom.clickbutton('Add zone', async function click() {
			let pcName = existingProviderConfigName.value;
			if (!pcName) {
				if (!fields) {
					alert('No provider selected.');
					return;
				}
				let pc = {
					Name: newProviderConfigName.value,
					ProviderName: providerName,
					ProviderConfigJSON: providerConfigJSON(fields),
				};
				pc = await check(fieldset, () => client.ProviderConfigAdd(pc));
				pcName = pc.Name;
			}
			const z = {
				Name: trimSuffix(zone.value, '.') + '.',
				ProviderConfigName: pcName,
				SerialLocal: 0,
				SerialRemote: 0,
				SyncInterval: 0,
				RefreshInterval: 0,
				NextSync: new Date(),
				NextRefresh: new Date(),
			};
			const nz = await check(fieldset, () => client.ZoneAdd(z, [])); // todo: allow specifying notifies
			zones.push(nz);
			render();
			close();
		}))))));
		zone.focus();
	})), dom.br(), dom.h1('Zones (Domains)'), dom.table(dom.thead(dom.tr(dom.th('Zone'), dom.th('Provider Config'), dom.th('Last sync'), dom.th('Last record change'), dom.th('Serial'), dom.th('Refresh next/interval'), dom.th('Sync next/interval'))), zonesTbody = dom.tbody()));
	const render = () => {
		const now = new Date();
		dom._kids(zonesTbody, zones.length ? [] : dom.tr(dom.td(attr.colspan('6'), 'No zones.', style({ textAlign: 'left' }))), zones.map(z => dom.tr(dom.td(dom.a(attr.href('#zones/' + trimDot(z.Name)), trimDot(z.Name))), dom.td(z.ProviderConfigName), dom.td(z.LastSync ? [formatAge(z.LastSync), attr.title(formatDate(z.LastSync))] : []), dom.td(z.LastRecordChange ? [formatAge(z.LastRecordChange), attr.title(formatDate(z.LastRecordChange))] : []), dom.td('' + z.SerialLocal, z.SerialLocal !== z.SerialRemote ? ' (at remote: ' + z.SerialRemote + ')' : '', attr.title(`Next SOA check in ${formatAge(undefined, z.NextRefresh)} at ${formatDate(z.NextRefresh)}.\nNext sync in ${formatAge(undefined, z.NextSync)} at ${formatDate(z.NextSync)}.`)), dom.td(formatAge(now, z.NextRefresh), ' / ', formatAge(now, new Date(now.getTime() + z.RefreshInterval / (1000 * 1000)))), dom.td(formatAge(now, z.NextSync), ' / ', formatAge(now, new Date(now.getTime() + z.SyncInterval / (1000 * 1000)))))));
	};
	render();
	return root;
};
// todo: add mechanims to keep age up to date while page is alive. with setInterval/setTimeout, and clearing those timers when we navigate away, like in ding. also use mechanism to keep propagation colors up to date.
const formatAge = (start, end) => {
	const second = 1;
	const minute = 60 * second;
	const hour = 60 * minute;
	const day = 24 * hour;
	const week = 7 * day;
	const year = 365 * day;
	const periods = [year, week, day, hour, minute, second];
	const suffix = ['y', 'w', 'd', 'h', 'm', 's'];
	if (!start) {
		start = new Date();
	}
	if (!end) {
		end = new Date();
	}
	let e = end.getTime() / 1000;
	let t = e - start.getTime() / 1000;
	let ago = false;
	if (t < 0) {
		t = -t;
		ago = true;
	}
	let s = '';
	for (let i = 0; i < periods.length; i++) {
		const p = periods[i];
		if (t >= 2 * p || i === periods.length - 1) {
			const n = Math.round(t / p);
			s = '' + n + suffix[i];
			break;
		}
	}
	if (ago) {
		s += ' ago';
	}
	return s;
};
const formatDate = (dt) => {
	return new Intl.DateTimeFormat(undefined, {
		weekday: 'short',
		year: 'numeric',
		month: 'short',
		day: 'numeric',
		hour: 'numeric',
		minute: 'numeric',
		second: 'numeric',
	}).format(dt);
};
const zoneRelName = (zone, s) => {
	s = s.substring(0, s.length - zone.Name.length);
	if (s) {
		s = s.substring(0, s.length - 1);
	}
	return s;
};
const box = (color) => dom.div(style({ width: '.75em', height: '.75em', display: 'inline-block', backgroundColor: color }));
const popupHistory = (absName, history) => {
	// todo: show as a visual timeline.
	const now = new Date();
	history = [...history];
	// Sort by last end time (still active record) first.
	history.sort((a, b) => {
		if (!a.End || !b.End) {
			return !a.End ? -1 : 1;
		}
		return b.End.getTime() - a.End.getTime();
	});
	const wildcardAbsName = '*.' + absName.substring(absName.indexOf('.') + 1);
	popup(dom.h1('History and propagation state'), dom.p('Previous versions or negative lookup results for this record set may still be in DNS resolver caches.'), dom.div(dom.div(box('#009fff'), ' Current record(s)'), dom.div(box('orange'), ' Previous records, potentially still in caches'), dom.div(box('#ffe300'), ' Negative lookup results potentially still in caches'), dom.div(box('#ccc'), ' Expired from caches')), dom.br(), dom.table(dom._class('striped'), dom.thead(dom.tr(dom.th('Status'), dom.th('Expiration'), dom.th('Since'), dom.th('Negative?'), dom.th('Wildcard?'), dom.th('TTL'), dom.th('Value(s)', style({ textAlign: 'left' })))), dom.tbody(history.map((state, index) => dom.tr(dom.td(box(!state.Negative && index === 0 && !state.Records[0].Deleted ? '#009fff' : (state.End && state.End.getTime() <= now.getTime() ? '#ccc' : (state.Negative ? '#ffe300' : 'orange')))), dom.td(!state.Negative && index === 0 && !state.Records[0].Deleted ? [] : [
		state.End.getTime() > now.getTime() ? ' in ' : '',
		formatAge(now, state.End),
		attr.title(formatDate(state.End)),
	]), dom.td(attr.title(formatAge(state.Start, now) + ' ago'), formatDate(state.Start)), dom.td(state.Negative ? 'Yes' : ''), dom.td(!state.Negative && state.Records[0].AbsName !== absName && state.Records[0].AbsName === wildcardAbsName ? 'Yes' : ''), dom.td(state.Negative ? [] : '' + state.Records[0].TTL), dom.td(style({ textAlign: 'left' }), state.Negative ? [] :
		dom.ul(style({ marginBottom: 0 }), (state.Records || []).map(r => dom.li(r.Value)))))))));
};
const popupEdit = (zone, records, isNew) => {
	let xtype;
	let relName;
	let ttl;
	let fieldset;
	// Meta types that we don't let the user create.
	const skipTypes = ['Reserved', 'None', 'NXNAME', 'OPT', 'UINFO', 'UID', 'GID', 'UNSPEC', 'TKEY', 'TSIG', 'IXFR', 'AXFR', 'MAILA', 'MAILB', 'ANY'];
	// Types we show first in the list. There is a long tail of uninteresting records.
	const firstTypes = ['A', 'AAAA', 'CAA', 'CNAME', 'DNSKEY', 'DS', 'MX', 'NS', 'OPENPGPKEY', 'PTR', 'SMIMEA', 'SOA', 'SRV', 'SSHFP', 'SVCB', 'TLSA', 'TXT'];
	const valuesView = {
		root: dom.div(),
		values: [],
	};
	const addValue = (s, recordID) => {
		const input = dom.input(attr.value(s), style({ width: '60em' }));
		let v;
		const root = dom.div(style({ display: 'flex', gap: '.5em', margin: '.5ex 0' }), input, dom.clickbutton('Remove', attr.title('Remove value'), function click() {
			valuesView.values.splice(valuesView.values.indexOf(v), 1);
			root.remove();
		}));
		v = { root: root, input: input, recordID: recordID };
		valuesView.values.push(v);
		valuesView.root.appendChild(v.root);
	};
	for (const r of records) {
		addValue(r.Value, r.ID);
	}
	if (records.length === 0) {
		addValue('', 0);
	}
	return new Promise((resolve, reject) => {
		const [close, closed] = popup(dom.h1(isNew ? 'New record set' : 'Edit record set'), dom.form(async function submit(e) {
			e.preventDefault();
			e.stopPropagation();
			const values = valuesView.values.map(vw => vw.input.value);
			if (values.length === 0) {
				alert('Specify at least one value.');
				throw new Error('must specify at least one value');
			}
			const n = {
				RelName: relName.value,
				Type: parseInt(xtype.value),
				TTL: parseInt(ttl.value),
				Values: values,
			};
			if (isNew) {
				await check(fieldset, () => client.RecordSetAdd(zone.Name, n));
			}
			else {
				await check(fieldset, () => client.RecordSetUpdate(zone.Name, zoneRelName(zone, records[0].AbsName), n, records.map(r => r.ID), valuesView.values.map(v => v.recordID)));
			}
			close();
			resolve();
		}, fieldset = dom.fieldset(style({ display: 'flex', flexDirection: 'column', gap: '2ex' }), dom.div(dom.label(dom.div('Name'), relName = dom.input(records.length > 0 ? attr.value(zoneRelName(zone, records[0].AbsName)) : [], style({ textAlign: 'right' })), '.' + zone.Name)), dom.div(dom.label(dom.div('TTL'), ttl = dom.input(attr.type('number'), attr.required(''), attr.value('300')))), dom.div(dom.label(dom.div('Type'), xtype = dom.select(dom.optgroup(attr.label('Common'), Object.entries(dnsTypeNames).filter(t => firstTypes.includes(t[1])).sort((ta, tb) => firstTypes.indexOf(ta[1]) - firstTypes.indexOf(tb[1])).map(t => dom.option(attr.value(t[0]), t[1]))), dom.optgroup(attr.label('Others'), Object.entries(dnsTypeNames).filter(t => !firstTypes.includes(t[1]) && !skipTypes.includes(t[1])).sort((ta, tb) => ta[1] < tb[1] ? -1 : 1).map(t => dom.option(attr.value(t[0]), t[1]))), isNew ? [] : attr.disabled(''), records.length === 0 ? [] : prop({ value: '' + records[0].Type })))), dom.div(dom.div(dom.label('Value(s)'), ' ', dom.clickbutton('Add', attr.title('Add another value'), function click() {
			addValue('', 0);
		})), valuesView), dom.div(dom.submitbutton(isNew ? 'Add record set' : 'Update record set')))));
		closed.then(undefined, reject);
		relName.focus();
	});
};
const pageZone = async (zonestr) => {
	let [zone, providerConfig, notifies0, credentials0, sets0] = await client.Zone(zonestr + '.');
	let notifies = notifies0 || [];
	let credentials = credentials0 || [];
	let sets = sets0 || [];
	dom._kids(crumbElem, dom.a(attr.href('#'), 'Home'), ' / ', dom.a(attr.href('#zones/' + trimDot(zone.Name)), 'Zone ' + trimDot(zone.Name)));
	document.title = 'Dnsclay - Zone ' + trimDot(zone.Name);
	const relName = (s) => zoneRelName(zone, s);
	const age = (r) => {
		let s = '';
		let title = 'First: ' + formatDate(r.First) + ' (serial ' + r.SerialFirst + ')\n';
		if (r.Deleted) {
			s = formatAge(r.Deleted);
			title += 'Deleted: ' + formatDate(r.Deleted) + ' (serial ' + r.SerialDeleted + ')\n';
		}
		else {
			s = formatAge(r.First);
		}
		return dom.span(s, attr.title(title));
	};
	let showHistoric;
	let showDNSSEC;
	let recordsTbody;
	const refresh = async (elem) => {
		const [nzone, npc, nnotifies, ncredentials, nsets] = await check(elem, () => client.Zone(zone.Name));
		zone = nzone;
		providerConfig = npc;
		notifies = nnotifies || [];
		credentials = ncredentials || [];
		sets = nsets || [];
		render();
	};
	const root = dom.div(dom.div(dom.p(`Provider config: ${zone.ProviderConfigName} (provider ${providerConfig.ProviderName})`), dom.clickbutton('Edit zone config', async function click(e) {
		let fieldset;
		let refreshival;
		let syncival;
		let providerConfigName;
		const providerConfigs = await check(e.target, () => client.ProviderConfigs()) || [];
		const [close] = popup(dom.h1('Edit zone'), dom.br(), dom.form(async function submit(e) {
			e.preventDefault();
			e.stopPropagation();
			const nz = { ...zone };
			nz.ProviderConfigName = providerConfigName.value;
			nz.RefreshInterval = 1000 * 1000 * 1000 * parseInt(refreshival.value);
			nz.SyncInterval = 1000 * 1000 * 1000 * parseInt(syncival.value);
			zone = await check(fieldset, () => client.ZoneUpdate(nz));
			close();
		}, fieldset = dom.fieldset(style({ display: 'flex', flexDirection: 'column', gap: '2ex' }), dom.label(dom.div('Refresh interval (in seconds)'), refreshival = dom.input(attr.type('number'), attr.required(''), attr.value('' + (zone.RefreshInterval / (1000 * 1000 * 1000))))), dom.label(dom.div('Sync interval (in seconds)'), syncival = dom.input(attr.type('number'), attr.required(''), attr.value('' + (zone.SyncInterval / (1000 * 1000 * 1000))))), dom.label(dom.div('Provider config'), providerConfigName = dom.select(providerConfigs.sort((a, b) => a.Name < b.Name ? -1 : 1).map(pc => dom.option(pc.Name)), prop({ value: zone.ProviderConfigName }))), dom.div(dom.submitbutton('Save')))));
	}), ' ', dom.clickbutton('Edit provider config', async function click(e) {
		let fieldset;
		const [stringEnums, providers] = await check(e.target, () => availableProviders());
		const p = providers.find(p => p.Name === 'Provider_' + providerConfig.ProviderName);
		if (!p) {
			alert('cannot find provider ' + providerConfig.ProviderName);
			return;
		}
		let testResult;
		let fields;
		const [close] = popup(dom.h1('Edit provider config'), dom.form(async function submit(e) {
			e.preventDefault();
			e.stopPropagation();
			// Test config.
			testResult.innerText = '';
			const nrecords = await check(fieldset, () => client.ProviderConfigTest(zone.Name, providerConfig.ProviderName, providerConfigJSON(fields)));
			testResult.innerText = 'Success, found ' + nrecords + ' DNS records';
		}, fieldset = dom.fieldset(style({ display: 'flex', flexDirection: 'column', gap: '2ex' }), dom.label(dom.div('Name'), dom.div(dom.input(attr.value(providerConfig.Name), attr.disabled('')))), dom.label(dom.div('Provider'), dom.div(dom.select(attr.disabled(''), dom.option(providerConfig.ProviderName), prop({ value: providerConfig.ProviderName })))), dom.div(style({ padding: '1em', border: '1px solid #ddd' }), dom.h2('Provider config'), dom.div(style({ display: 'flex', flexDirection: 'column', gap: '2ex' }), fields = providerFields(p, stringEnums, providerConfig.ProviderConfigJSON))), dom.div(dom.submitbutton('Test config'), ' ', testResult = dom.span()), dom.div(dom.clickbutton('Save', async function click() {
			let npc = {
				Name: providerConfig.Name,
				ProviderName: providerConfig.ProviderName,
				ProviderConfigJSON: providerConfigJSON(fields),
			};
			providerConfig = await check(fieldset, () => client.ProviderConfigUpdate(npc));
			close();
		})))));
	})), dom.br(), dom.div(style({ display: 'flex', gap: '1em' }), dom.div(style({ backgroundColor: '#f4f4f4', border: '1px solid #ddd', borderRadius: '.25em', padding: '.5em' }), dom.div(style({ display: 'flex', gap: '.5em', alignItems: 'baseline' }), dom.h2('DNS NOTIFY addresses'), dom.clickbutton('Add', function click() {
		let address;
		let fieldset;
		const [close] = popup(dom.h1('Add DNS NOTIFY address'), dom.form(async function submit(e) {
			e.preventDefault();
			e.stopPropagation();
			let zn = {
				ID: 0,
				Created: new Date(),
				Zone: zone.Name,
				Protocol: fieldset.querySelector('input[name=notifyprotocol]:checked')?.value || '',
				Address: address.value,
			};
			const nzn = await check(fieldset, () => client.ZoneNotifyAdd(zn));
			notifies.push(nzn);
			close();
			location.reload(); // todo: render the list again
		}, fieldset = dom.fieldset(style({ display: 'flex', flexDirection: 'column', gap: '2ex' }), dom.div(dom.div(dom.label('Protocol')), dom.label(dom.input(attr.type('radio'), attr.name('notifyprotocol'), attr.value('tcp')), ' tcp'), ' ', dom.label(dom.input(attr.type('radio'), attr.name('notifyprotocol'), attr.value('udp')), ' udp')), dom.div(dom.div(dom.label('Address')), address = dom.input(attr.type('required'), attr.placeholder('127.0.0.1:53'))), dom.div(dom.submitbutton('Add')))));
	})), dom.table(dom.thead(dom.tr(dom.th('Protocol'), dom.th('Address'), dom.th())), dom.tbody(notifies.length ? [] : dom.tr(dom.td(attr.colspan('3'), 'No notify addressses.', style({ textAlign: 'left' }))), notifies.map(n => {
		const row = dom.tr(dom.td(n.Protocol), dom.td(n.Address), dom.td(dom.clickbutton('Notify', async function click(e) {
			await check(e.target, () => client.ZoneNotify(n.ID));
		}), ' ', dom.clickbutton('Delete', async function click(e) {
			if (!confirm('Are you sure?')) {
				return;
			}
			await check(e.target, () => client.ZoneNotifyDelete(n.ID));
			notifies.splice(notifies.indexOf(n), 1);
			row.remove();
		})));
		return row;
	})))), dom.div(style({ backgroundColor: '#f4f4f4', border: '1px solid #ddd', borderRadius: '.25em', padding: '.5em' }), dom.div(style({ display: 'flex', gap: '.5em', alignItems: 'baseline' }), dom.h2('Credentials'), ' ', dom.clickbutton('Add', function click() {
		let name;
		let key;
		let fieldset;
		const [close] = popup(dom.h1('Add credential'), dom.p('For use with DNS UPDATE and DNS AXFR/IXFR.'), dom.form(async function submit(e) {
			e.preventDefault();
			e.stopPropagation();
			const typ = fieldset.querySelector('input[name=credentialtype]:checked')?.value || '';
			let c = {
				ID: 0,
				Created: new Date(),
				Name: name.value,
				Type: typ,
				TSIGSecret: typ === 'tsig' ? key.value : '',
				TLSPublicKey: typ === 'tlspubkey' ? key.value : '',
			};
			const nc = await check(fieldset, () => client.ZoneCredentialAdd(zone.Name, c));
			credentials.push(nc);
			close();
			location.reload(); // todo: render the list again
		}, fieldset = dom.fieldset(style({ display: 'flex', flexDirection: 'column', gap: '2ex' }), dom.div(dom.div(dom.label('Name')), name = dom.input(attr.type('required'), attr.placeholder('name-with-dashes-or-dots'), style({ width: '100%' })), dom.div(style({ fontStyle: 'italic' }), 'Must be a valid DNS name for TSIG.')), dom.div(dom.div(dom.label('Type')), dom.label(dom.input(attr.type('radio'), attr.name('credentialtype'), attr.value('tsig')), ' TSIG'), ' ', dom.label(dom.input(attr.type('radio'), attr.name('credentialtype'), attr.value('tlspubkey')), ' TLS public key')), dom.div(dom.div(dom.label('TSIG secret or TLS public key')), key = dom.input(style({ width: '100%' })), dom.div(style({ fontStyle: 'italic' }), 'In case of a TSIG secret, if left empty, a random key will be generated.')), dom.div(dom.submitbutton('Add')))));
	})), dom.table(dom.thead(dom.tr(dom.th('Name'), dom.th('Type'), dom.th('TSIG Secret / TLS public key'), dom.th('Age'), dom.th(''))), dom.tbody(credentials.length ? [] : dom.tr(dom.td(attr.colspan('5'), 'No credentials.', style({ textAlign: 'left' }))), credentials.map(c => {
		const row = dom.tr(dom.td(c.Name), dom.td(c.Type), dom.td(c.Type === 'tsig' ?
			dom.clickbutton('Show', function click(e) {
				e.target.replaceWith(dom.span(c.TSIGSecret));
			}) : c.TLSPublicKey), dom.td(formatAge(c.Created), attr.title(formatDate(c.Created))), dom.td(dom.clickbutton('Delete', async function click(e) {
			if (!confirm('Are you sure?')) {
				return;
			}
			await check(e.target, () => client.ZoneCredentialDelete(c.ID));
			credentials.splice(credentials.indexOf(c), 1);
			row.remove();
		})));
		return row;
	}))))), dom.br(), dom.div(style({ display: 'flex', gap: '.5em', alignItems: 'baseline' }), dom.h2('Records'), ' ', dom.clickbutton('Add records', async function click(e) {
		await popupEdit(zone, [], true);
		await refresh(e.target);
	}), ' ', dom.clickbutton('Import records', attr.title('Import records from zone file'), function click() {
		let zonefile;
		let fieldset;
		const [close] = popup(dom.h1('Import records from zone file'), dom.form(async function submit(e) {
			e.preventDefault();
			e.stopPropagation();
			await check(fieldset, () => client.ZoneImportRecords(zone.Name, zonefile.value));
			await refresh(fieldset);
			close();
		}, fieldset = dom.fieldset(style({ display: 'flex', flexDirection: 'column', gap: '2ex' }), dom.div(dom.label(dom.div('Zone file'), zonefile = dom.textarea('$TTL 300 ; default 5m\n$ORIGIN ' + zone.Name + '\n\n; record syntax: name ttl type value\n; example:\n;relativename 300 A 1.2.3.4\n\n', style({ width: '60em' }), attr.rows('10')))), dom.div(dom.submitbutton('Import')))));
		zonefile.focus();
	}), ' ', dom.clickbutton('Fetch latest records', async function click(e) {
		const [_, nsets] = await check(e.target, () => client.ZoneRefresh(zone.Name));
		sets = nsets || [];
		render();
	}), ' ', dom.clickbutton('Purge history', attr.title('Remove history with previously existing but now removed records. History is used by IXFR for incremental zone transfers, but IXFR attempts will fall back to AXFR if history is not available.'), async function click(e) {
		if (!confirm('Are you sure?')) {
			return;
		}
		const [_, nsets] = await check(e.target, () => client.ZonePurgeHistory(zone.Name));
		sets = nsets || [];
		render();
	}), ' '), dom.div(dom.label(showHistoric = dom.input(attr.type('checkbox'), localStorage.getItem('showHistoric') === 'yes' ? attr.checked('') : [], function change() {
		if (showHistoric.checked) {
			localStorage.setItem('showHistoric', 'yes');
		}
		else {
			localStorage.removeItem('showHistoric');
		}
		render();
	}), ' Show historic records'), ' ', dom.label(showDNSSEC = dom.input(attr.type('checkbox'), localStorage.getItem('showDNSSEC') === 'yes' ? attr.checked('') : [], function change() {
		if (showDNSSEC.checked) {
			localStorage.setItem('showDNSSEC', 'yes');
		}
		else {
			localStorage.removeItem('showDNSSEC');
		}
		render();
	}), ' Show DNSSEC signature records', attr.title('RRSIG, NSEC and NSEC3 records are hidden by default'))), dom.table(dom._class('hover'), dom._class('striped'), dom.thead(dom.tr(dom.th(), dom.th('Age'), dom.th('Name'), dom.th('TTL'), dom.th('Type'), dom.th('Value'), dom.th('Actions'))), recordsTbody = dom.tbody()), dom.br(), dom.h2('Danger'), dom.clickbutton('Remove zone', attr.title('Remove zone from management in dnsclay. The zone and its records are not changed at the provider.'), async function click(e) {
		if (!confirm('Are you sure you want to remove this zone from management in dnsclay? The zone and its records are not changed at the provider.')) {
			return;
		}
		await check(e.target, () => client.ZoneDelete(zone.Name));
		location.hash = '#';
	}));
	const render = () => {
		// todo: implement sorting on other columns. at least type.
		sets = sets.sort((sa, sb) => {
			const a = sa.Records[0];
			const b = sb.Records[0];
			if (a.AbsName !== b.AbsName) {
				return a.AbsName.split('.').reverse().join('.') < b.AbsName.split('.').reverse().join('.') ? -1 : 1;
			}
			const ta = dnsTypeNames[a.Type] || ('' + a.Type);
			const tb = dnsTypeNames[b.Type] || ('' + b.Type);
			if (ta !== tb) {
				return ta < tb ? -1 : 1;
			}
			const da = !!a.Deleted;
			const db = !!b.Deleted;
			if (da !== db) {
				return da ? 1 : -1;
			}
			if (!a.Deleted && a.Value !== b.Value) {
				return a.Value < b.Value ? -1 : 1;
			}
			let tma = a.Deleted || a.First;
			let tmb = b.Deleted || b.First;
			return tma.getTime() > tmb.getTime() ? -1 : 1;
		});
		// todo: add checkboxes to select multiple/all records and do mass operations like changing TTL and deleting them.
		// todo: add mechanism to show history for an rrset (name + type). possibly a popup
		const typeRRSIG = 46;
		const typeNSEC = 47;
		const typeNSEC3 = 50;
		dom._kids(recordsTbody, sets
			.filter(s => (showHistoric.checked || !s.Records[0].Deleted) && (showDNSSEC.checked || (s.Records[0].Type !== typeRRSIG && s.Records[0].Type !== typeNSEC && s.Records[0].Type !== typeNSEC3)))
			.map(set => {
			const r0 = set.Records[0];
			const hasNegative = !!(set.States || []).find(state => state.Negative);
			const hasPrevious = !!(set.States || []).find(state => !state.Negative);
			let propagationText = [];
			if (hasPrevious) {
				propagationText.push('Due to TTLs, previous versions of this record may still be cached in resolvers.');
			}
			if (hasNegative) {
				propagationText.push('Due to the TTL for lookups with negative result, absence of this record may still be cached in resolvers.');
			}
			return dom.tr(r0.Deleted ? [style({ color: '#888' }), attr.title('Historic/deleted record')] : [], dom.td(hasNegative && hasPrevious ? style({ backgroundColor: 'orange', border: '3px solid #ffe300' }) : [], hasNegative && !hasPrevious ? style({ backgroundColor: '#ffe300' }) : [], !hasNegative && hasPrevious ? style({ backgroundColor: 'orange' }) : [], (hasNegative || hasPrevious) ? [] : style({ color: '#888' }), propagationText.length > 0 ? attr.title(propagationText.join('\n')) : []), dom.td(age(r0)), dom.td(relName(r0.AbsName), style({ textAlign: 'right' })), dom.td('' + r0.TTL), dom.td(dnsTypeNames[r0.Type] || ('' + r0.Type), attr.title('Type ' + r0.Type + (r0.ProviderID ? '\nID at provider: ' + r0.ProviderID : '')), style({ textAlign: 'left' })), dom.td(style({ textAlign: 'left' }), (set.Records || []).map(r => dom.div(style({ wordBreak: 'break-all' }), r.Value))), dom.td(style({ whiteSpace: 'nowrap' }), r0.Deleted ?
				dom.clickbutton('Recreate', async function click(e) {
					await popupEdit(zone, set.Records || [], true);
					await refresh(e.target);
				}) : [
				dom.clickbutton('Edit', async function click(e) {
					await popupEdit(zone, set.Records || [], false);
					await refresh(e.target);
				}), ' ',
				dom.clickbutton('Delete', async function click(e) {
					if (!confirm('Are you sure you want to delete this record set?')) {
						return;
					}
					await check(e.target, () => client.RecordSetDelete(zone.Name, relName(r0.AbsName), r0.Type, set.Records.map(r => r.ID)));
					await refresh(e.target);
				}),
			], ' ', dom.clickbutton('History', async function click(e) {
				const hist = await check(e.target, () => client.ZoneRecordSetHistory(zone.Name, relName(r0.AbsName), r0.Type));
				popupHistory(r0.AbsName, hist || []);
			})));
		}));
	};
	render();
	return root;
};
const hashchange = async (e) => {
	const hash = decodeURIComponent(window.location.hash.substring(1));
	const t = hash.split('/');
	try {
		let elem;
		if (t.length === 1 && t[0] === '') {
			elem = await pageHome();
		}
		else if (t.length === 2 && t[0] === 'zones') {
			elem = await pageZone(t[1]);
		}
		else {
			window.alert('Unknown hash');
			location.hash = '#';
			return;
		}
		dom._kids(pageElem, elem);
	}
	catch (err) {
		window.alert('Error: ' + err.message);
		window.location.hash = e?.oldURL ? new URL(e.oldURL).hash : '';
		throw err;
	}
};
const init = async () => {
	[version, dnsTypeNames] = await Promise.all([
		client.Version(),
		client.DNSTypeNames(),
	]);
	const root = dom.div(dom.div(style({ display: 'flex', justifyContent: 'space-between', marginBottom: '1ex', padding: '.5em 1em', backgroundColor: '#f8f8f8' }), crumbElem, dom.div(dom.a(attr.href('https://github.com/mjl-/dnsclay'), 'dnsclay'), ' ', version, ' ', dom.a(attr.href('license'), 'license'))), dom.div(pageElem));
	document.getElementById('rootElem').replaceWith(root);
	rootElem = root;
	window.addEventListener('hashchange', hashchange);
	await hashchange();
};
window.addEventListener('load', async () => {
	try {
		await init();
	}
	catch (err) {
		window.alert('Error: ' + err.message);
	}
});
