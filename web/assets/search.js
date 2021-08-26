const fusePromise = (async () => {
    const options = {
        threshold: 0.5,
        includeMatches: true,
        findAllMatches: true,
        useExtendedSearch: true,
        keys: ["name"]
    };

    const res = await window.fetch('/search/terms.json')
    const data = await res.json()
    return new Fuse(data, options)
})()

const search = document.getElementById('search')
const resultsContainer = document.getElementById('search_results')

function containsCharactersInOrder(name, query) {
    for (ch of name) {
        if (!query) {
            return true
        }

        if (ch.match(/\s/)) {
            continue
        }

        if (query[0].match(/\s/) || query[0].localeCompare(
            ch,
            undefined,
            { sensitivity: 'accent' },
        ) === 0) {
            query = query.substr(1)
        }
    }

    return !query;
}

function computeMatchingRegions(name, query) {
    name = name.toLowerCase()
    query = query.toLowerCase()

    const indices = []

    let nameIndex = 0
    let matchIndex = 0
    let withinMatch = false

    while (name && query) {
        const queryChar = query[0]
        if (queryChar.match(/\s/)) {
            query = query.substr(1)
            continue
        }

        const nameChar = name[0]
        nameIndex++
        name = name.substr(1)

        if (nameChar.match(/\s/)) {
            continue
        }

        if (!withinMatch) {
            const i = name.indexOf(query)
            if (i !== -1) {
                indices.push([nameIndex+i, nameIndex+i + query.length - 1])
                return indices
            }
        }

        if (nameChar === queryChar) {
            query = query.substr(1)    
            
            if (!withinMatch) {
                withinMatch = true
                matchIndex = nameIndex - 1
            }
        } else if (withinMatch) {
            withinMatch = false
            indices.push([matchIndex, nameIndex - 2])
        }
    }

    if (query) {
        return [];
    }

    if (withinMatch) {
        indices.push([matchIndex, nameIndex - 1])
    }

    return indices;
}

search.addEventListener("search", async (event) => {
    if (search.value.length === 0) {
        resultsContainer.style.display = 'none'
        return
    }

    const query = search.value
    const fuse = await fusePromise
    const results = fuse.search(query)
    const card = resultsContainer.getElementsByClassName("card-body")[0]

    card.innerHTML = ''
    
    for (const r of results) {
        const indices = computeMatchingRegions(r.item.name, query)

        if (indices.length === 0) {
            continue
        }

        const container = document.createElement('p')
        container.className = 'result'

        const link = document.createElement('a')
        link.setAttribute("href", r.item.uri)

        let prev = 0
        for (const [begin, end] of indices) {
            if (begin > prev) {
                const text = r.item.name.substring(prev, begin)
                link.append(text)
            }

            const elem = document.createElement('mark')
            elem.innerText = r.item.name.substring(begin, end + 1)
            link.appendChild(elem)
            prev = end + 1
        }

        // for (const match of r.matches) {
        //     for (const [begin, end] of match.indices) {
        //         if (begin > prev) {
        //             const text = r.item.name.substring(prev, begin)
        //             link.append(text)
        //         }

        //         if (begin < prev) {
        //             // Bug in fuse.js?
        //             break
        //         }

        //         const elem = document.createElement('mark')
        //         elem.innerText = r.item.name.substring(begin, end + 1)
        //         link.appendChild(elem)
        //         prev = end + 1
        //     }
        // }

        if (r.item.name.length > prev) {
            const text = r.item.name.substr(prev)
            link.append(text)
        }

        container.appendChild(link)
        container.append(' ')

        const type = document.createElement('span')
        const sticky = document.createElement('span')
        sticky.className = 'sticky-note'

        switch (r.item.type) {
            case "command":
            case "event":
            case "timeout":
                type.className = 'role-' + r.item.type
                type.appendChild(sticky)
                type.append(" ")
                break;

            case "aggregate":
            case "process":
            case "projection":
            case "integration":
                type.className = 'handlertype-' + r.item.type
                type.appendChild(sticky)
                type.append(" ")
                break;
        }

        type.append(r.item.type)
        container.appendChild(type)

        if (r.item.docs) {
            const docs = document.createElement('div')
            docs.className = 'docs'
            
            const text = document.createTextNode(r.item.docs)
            docs.appendChild(text)

            container.appendChild(docs)
        }

        card.appendChild(container)
    }

    if (!card.innerHTML) {
        const message = document.createElement('span')
        message.className = 'text-muted'
        message.innerText = "No results match this query."
        card.appendChild(message)
    }

    resultsContainer.style.display = 'block'
});
